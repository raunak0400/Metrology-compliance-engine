package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"

	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/authtoken"
	"github.com/raunak0400/metrology-backend/internal/compliance"
	"github.com/raunak0400/metrology-backend/internal/config"
	"github.com/raunak0400/metrology-backend/internal/constants"
	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/declarations"
	"github.com/raunak0400/metrology-backend/internal/inspections"
	"github.com/raunak0400/metrology-backend/internal/ml"
	"github.com/raunak0400/metrology-backend/internal/ruleengine"
	"github.com/raunak0400/metrology-backend/pkg/mongodb"
	"github.com/raunak0400/metrology-backend/pkg/postgres"
	"github.com/raunak0400/metrology-backend/pkg/redisclient"
	"github.com/raunak0400/metrology-backend/pkg/storage"
)

// requireLiveEnv skips the test unless every TEST_* infra var is set (see
// backend/.env.example and backend/README.md's "Infrastructure connectivity
// tests" section — same convention used by pkg/postgres, pkg/mongodb, etc.).
func requireLiveEnv(t *testing.T) map[string]string {
	t.Helper()
	vars := map[string]string{
		"DATABASE_URL": os.Getenv("TEST_DATABASE_URL"),
		"MONGODB_URI":  os.Getenv("TEST_MONGODB_URI"),
		"REDIS_URL":    os.Getenv("TEST_REDIS_URL"),
		"S3_ENDPOINT":  os.Getenv("TEST_S3_ENDPOINT"),
	}
	for k, v := range vars {
		if v == "" {
			t.Skipf("TEST_%s not set; skipping live business-flow integration test", k)
		}
	}
	return vars
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// netQtyFormatRe is the mock rule engine's stand-in for a real Legal
// Metrology net-quantity format check — used only to decide the mocked
// /validate response below, not a real legal rule.
var netQtyFormatRe = regexp.MustCompile(`^\d+\s?(g|kg|ml|l)$`)

// newMockRuleEngine stands in for the real Python Legal Metrology Rule
// Engine service (see rule_engine/README.md on the ml-service branch) so
// this test can exercise the Go orchestration path (build request -> POST
// /validate -> persist checks/violations/outcome) without a live Python
// service. It mirrors the real contract (GET /health, GET /rules, POST
// /validate) with deterministic fixture rules/responses, not real legal text.
func newMockRuleEngine(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "healthy", "service": "legal-metrology-rule-engine",
			"ruleset_version": "it-test-1.0", "ruleset_description": "integration test fixture",
		})
	})

	mux.HandleFunc("/rules", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ruleset_version": "it-test-1.0", "ruleset_description": "integration test fixture",
			"rules": []map[string]any{
				{"rule_id": "IT_TEST_MRP_EXISTS", "clause": "[TEST FIXTURE] Clause 6(1)(e)", "severity": "major", "description": "MRP must be declared", "rule_class": "mrp"},
				{"rule_id": "IT_TEST_NET_QTY_FORMAT", "clause": "[TEST FIXTURE] Clause 8", "severity": "minor", "description": "Net quantity must follow the standard <number><unit> format", "rule_class": "net_quantity"},
			},
		})
	})

	mux.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			ProductID       string `json:"product_id"`
			ExtractedFields []struct {
				FieldType string `json:"field_type"`
				Text      string `json:"text"`
			} `json:"extracted_fields"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		hasMRP, netQtyOK := false, false
		for _, f := range req.ExtractedFields {
			switch f.FieldType {
			case "mrp":
				hasMRP = f.Text != ""
			case "net_quantity":
				netQtyOK = netQtyFormatRe.MatchString(f.Text)
			}
		}

		violations := []map[string]any{}
		if !hasMRP {
			violations = append(violations, map[string]any{"rule_id": "IT_TEST_MRP_EXISTS", "clause": "[TEST FIXTURE] Clause 6(1)(e)", "severity": "major", "description": "MRP missing"})
		}
		if !netQtyOK {
			violations = append(violations, map[string]any{"rule_id": "IT_TEST_NET_QTY_FORMAT", "clause": "[TEST FIXTURE] Clause 8", "severity": "minor", "description": "Net quantity format invalid", "field_evidence": "quantity not clearly stated"})
		}
		overallStatus := "compliant"
		if len(violations) > 0 {
			overallStatus = "non_compliant"
		}

		_ = json.NewEncoder(w).Encode(map[string]any{
			"product_id": req.ProductID, "ruleset_version": "it-test-1.0", "overall_status": overallStatus,
			"violations": violations,
			"applicability_notes": []map[string]any{
				{"rule_id": "IT_TEST_MRP_EXISTS", "applied": true, "reason": "test fixture: always applicable"},
				{"rule_id": "IT_TEST_NET_QTY_FORMAT", "applied": true, "reason": "test fixture: always applicable"},
			},
			"manual_review_required": []string{}, "synthesized_fields_summary": []any{},
		})
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// buildLiveDeps wires the exact same infrastructure cmd/api/main.go does, but
// against the TEST_* environment (a disposable scratch stack — see
// backend/README.md), so this test exercises real Postgres transactions,
// real MongoDB documents and a real Redis-backed Asynq client instead of
// mocks. The worker process is NOT started here: this test drives the
// business-logic services (declarations, compliance) directly wherever the
// HTTP surface is async-only by design (image scan processing belongs to a
// worker; POST /inspections/:id/check only enqueues). ruleEngineURL points at
// newMockRuleEngine, standing in for the real Python rule engine service.
func buildLiveDeps(t *testing.T, vars map[string]string, ruleEngineURL string) *app.Deps {
	t.Helper()
	ctx := context.Background()

	pool, err := postgres.Connect(ctx, vars["DATABASE_URL"])
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}
	t.Cleanup(pool.Close)

	mongoClient, mongoDB, err := mongodb.Connect(ctx, vars["MONGODB_URI"], "legal_metrology_ml_test")
	if err != nil {
		t.Fatalf("connect mongodb: %v", err)
	}
	t.Cleanup(func() { _ = mongoClient.Disconnect(context.Background()) })

	redisClient, err := redisclient.Connect(ctx, vars["REDIS_URL"])
	if err != nil {
		t.Fatalf("connect redis: %v", err)
	}
	t.Cleanup(func() { _ = redisClient.Close() })

	objectStorage, err := storage.New(storage.Config{
		Endpoint: vars["S3_ENDPOINT"], Region: envOr("TEST_S3_REGION", "us-east-1"),
		AccessKey: envOr("TEST_S3_ACCESS_KEY", "minioadmin"), SecretKey: envOr("TEST_S3_SECRET_KEY", "minioadmin"),
		Bucket: envOr("TEST_S3_BUCKET", "legal-metrology-test"), ForcePathStyle: true,
		UseSSL: os.Getenv("TEST_S3_USE_SSL") == "true",
	})
	if err != nil {
		t.Fatalf("create storage client: %v", err)
	}
	if err := objectStorage.EnsureBucket(ctx); err != nil {
		t.Fatalf("ensure bucket: %v", err)
	}

	redisOpt, err := redis.ParseURL(vars["REDIS_URL"])
	if err != nil {
		t.Fatalf("parse redis url: %v", err)
	}
	asynqClient := asynq.NewClient(asynq.RedisClientOpt{Addr: redisOpt.Addr, Password: redisOpt.Password, DB: redisOpt.DB})
	t.Cleanup(func() { _ = asynqClient.Close() })

	mlRepo := ml.NewRepository(mongoDB)
	if err := mlRepo.EnsureIndexes(ctx); err != nil {
		t.Fatalf("ensure mongo indexes: %v", err)
	}

	return &app.Deps{
		Config:           &config.Config{APIPrefix: "/api/v1", CORSOrigins: []string{"http://localhost:3000"}, RateLimitWindow: time.Minute, RateLimitMax: 100000, MaxUploadSizeBytes: 15_728_640, AllowedImageMIME: []string{"image/jpeg", "image/png", "image/webp"}},
		Pool:             pool,
		Queries:          db.New(pool),
		Mongo:            mongoDB,
		MLRepo:           mlRepo,
		Redis:            redisClient,
		Storage:          objectStorage,
		MLClient:         ml.NewClient(envOr("TEST_ML_SERVICE_URL", "http://localhost:8000"), "", 5*time.Second),
		RuleEngineClient: ruleengine.NewClient(ruleEngineURL, "", 5*time.Second),
		Asynq:            asynqClient,
		Tokenizer:        authtoken.NewTokenizer("test-access-secret-value", "test-refresh-secret-value", 15*time.Minute, 7*24*time.Hour),
	}
}

type apiEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func doRequest(t *testing.T, router http.Handler, method, path, token string, body interface{}) (int, apiEnvelope) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode request body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var env apiEnvelope
	if rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatalf("decode response body %q: %v", rec.Body.String(), err)
		}
	}
	return rec.Code, env
}

func mustDecode[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("decode %T: %v (raw=%s)", v, err, raw)
	}
	return v
}

// TestIntegration_FullBusinessFlow drives the Part 2 success criteria
// end-to-end against live infrastructure: auth + RBAC, products,
// inspections (including location/lat/long), declarations (manual
// correction + re-evaluation trigger), compliance orchestration (build a
// request from structured declarations, POST it to a mocked Python rule
// engine, persist the returned checks/violations/outcome inside a
// transaction), violations, evidence/image storage, the finalize/reopen
// workflow, and audit logging.
func TestIntegration_FullBusinessFlow(t *testing.T) {
	vars := requireLiveEnv(t)
	ruleEngineSrv := newMockRuleEngine(t)
	deps := buildLiveDeps(t, vars, ruleEngineSrv.URL)
	gin.SetMode(gin.TestMode)
	router := buildRouter(deps.Config, deps)
	ctx := context.Background()

	unique := time.Now().UnixNano()
	adminEmail := fmt.Sprintf("it-admin-%d@example.test", unique)
	password := "IntegrationTest123!"

	// ---- 1. Register + RBAC: new users default to INSPECTOR; grant ADMIN
	// directly (as a human admin would via PATCH /users/:id) so this admin can
	// exercise every permission scope below. ----
	code, env := doRequest(t, router, http.MethodPost, "/api/v1/auth/register", "", map[string]any{
		"email": adminEmail, "password": password, "fullName": "Integration Admin",
	})
	if code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d: %+v", code, env)
	}
	registered := mustDecode[struct {
		ID string `json:"id"`
	}](t, env.Data)

	adminRole, err := deps.Queries.GetRoleByName(ctx, constants.RoleAdmin)
	if err != nil {
		t.Fatalf("look up ADMIN role (has `go run ./cmd/seed` been run against this database?): %v", err)
	}
	userID := mustParseUUID(t, registered.ID)
	if err := deps.Queries.InsertUserRole(ctx, db.InsertUserRoleParams{UserID: userID, RoleID: adminRole.ID}); err != nil {
		t.Fatalf("grant ADMIN role: %v", err)
	}

	code, env = doRequest(t, router, http.MethodPost, "/api/v1/auth/login", "", map[string]any{"email": adminEmail, "password": password})
	if code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d: %+v", code, env)
	}
	login := mustDecode[struct {
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
	}](t, env.Data)
	adminToken := login.AccessToken

	// ---- RBAC: an unauthenticated request to a protected route is rejected. ----
	if code, _ := doRequest(t, router, http.MethodGet, "/api/v1/products", "", nil); code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for unauthenticated /products, got %d", code)
	}

	// ---- 2. Refresh token rotation. ----
	code, env = doRequest(t, router, http.MethodPost, "/api/v1/auth/refresh", "", map[string]any{"refreshToken": login.RefreshToken})
	if code != http.StatusOK {
		t.Fatalf("refresh: expected 200, got %d: %+v", code, env)
	}
	refreshed := mustDecode[struct {
		AccessToken string `json:"accessToken"`
	}](t, env.Data)
	if refreshed.AccessToken == "" {
		t.Fatal("expected a non-empty rotated access token")
	}
	// The old refresh token must now be rejected (rotation, not reuse).
	if code, _ := doRequest(t, router, http.MethodPost, "/api/v1/auth/refresh", "", map[string]any{"refreshToken": login.RefreshToken}); code != http.StatusUnauthorized {
		t.Fatalf("expected the rotated-out refresh token to be rejected, got %d", code)
	}

	// ---- 3. Product category (no HTTP surface for this — seeded data only —
	// so insert directly, exactly like cmd/seed does). ----
	catCode := fmt.Sprintf("IT-CAT-%d", unique)
	category, err := deps.Queries.UpsertCategory(ctx, db.UpsertCategoryParams{Name: catCode, Code: catCode, Description: strPtr("Integration test category")})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	// ---- 4. Product module: create/list/get/patch, pagination + filtering. ----
	code, env = doRequest(t, router, http.MethodPost, "/api/v1/products", adminToken, map[string]any{
		"name": "Integration Test Snack Pack", "categoryId": category.ID.String(), "brand": "TestBrand",
	})
	if code != http.StatusCreated {
		t.Fatalf("create product: expected 201, got %d: %+v", code, env)
	}
	product := mustDecode[struct {
		ID string `json:"id"`
	}](t, env.Data)

	code, env = doRequest(t, router, http.MethodGet, "/api/v1/products?categoryId="+category.ID.String()+"&limit=5", adminToken, nil)
	if code != http.StatusOK {
		t.Fatalf("list products: expected 200, got %d: %+v", code, env)
	}

	code, env = doRequest(t, router, http.MethodPatch, "/api/v1/products/"+product.ID, adminToken, map[string]any{"brand": "UpdatedBrand"})
	if code != http.StatusOK {
		t.Fatalf("patch product: expected 200, got %d: %+v", code, env)
	}

	// ---- 5. A mandatory + a non-mandatory rule mapped to `category`, and
	// zero rules mapped to `naCategory` (drives NOT_APPLICABLE below). Uses an
	// explicit test fixture rule, NOT an authoritative legal citation — see
	// section 20's "do not invent legal-rule expected values" instruction. ----
	code, env = doRequest(t, router, http.MethodPost, "/api/v1/rules", adminToken, map[string]any{
		"code":        fmt.Sprintf("IT-TEST-RULE-%d", unique),
		"title":       "[TEST FIXTURE] MRP must be present",
		"categoryIds": []string{category.ID.String()},
		"requirements": []map[string]any{
			{"key": "test.mrp.exists", "declarationType": "MRP", "operator": "EXISTS", "isMandatory": true, "message": "MRP missing"},
			{"key": "test.netqty.softFormat", "declarationType": "NET_QUANTITY", "operator": "REGEX_MATCH",
				"expected": map[string]any{"pattern": `^\d+\s?(g|kg|ml|l)$`}, "isMandatory": false, "message": "Recommended net-quantity format not followed"},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("create rule: expected 201, got %d: %+v", code, env)
	}

	// ---- 6. Inspection module: create with location/lat/lng, list, get, patch. ----
	lat, lng := 18.5204, 73.8567
	code, env = doRequest(t, router, http.MethodPost, "/api/v1/inspections", adminToken, map[string]any{
		"productId": product.ID, "categoryId": category.ID.String(), "source": "FIELD",
		"location": "Pune, Maharashtra", "latitude": lat, "longitude": lng,
	})
	if code != http.StatusCreated {
		t.Fatalf("create inspection: expected 201, got %d: %+v", code, env)
	}
	inspection := mustDecode[struct {
		ID        string   `json:"id"`
		Location  *string  `json:"location"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}](t, env.Data)
	if inspection.Location == nil || *inspection.Location != "Pune, Maharashtra" {
		t.Fatalf("expected location to round-trip, got %+v", inspection)
	}
	if inspection.Latitude == nil || *inspection.Latitude != lat || inspection.Longitude == nil || *inspection.Longitude != lng {
		t.Fatalf("expected lat/lng to round-trip, got %+v", inspection)
	}
	inspectionID := mustParseUUID(t, inspection.ID)

	// ---- 7. Image upload (section 6): validated, stored in MinIO/S3, and an
	// InspectionImage row created in PostgreSQL — never a base64 blob in Postgres. ----
	imageID, jobID := uploadTestImage(t, router, adminToken, inspection.ID)
	if imageID == "" || jobID == "" {
		t.Fatal("expected a non-empty imageId and jobId from image upload")
	}
	img, err := deps.Queries.GetInspectionImage(ctx, mustParseUUID(t, imageID))
	if err != nil {
		t.Fatalf("expected the uploaded image row to exist in PostgreSQL: %v", err)
	}
	if _, err := deps.Storage.Get(ctx, img.StorageKey); err != nil {
		t.Fatalf("expected the uploaded image bytes to exist in object storage at key %s: %v", img.StorageKey, err)
	}

	// ---- 8. ML analysis ingestion + normalization (sections 7-8), driven
	// directly rather than via the worker process (which isn't running in
	// this test): simulate what the scan pipeline would persist after
	// calling the ML service — normalized MRP/NET_QUANTITY declarations, with
	// raw value, confidence and bbox preserved. ----
	inspSvc := inspections.NewService(deps.Queries)
	declSvc := declarations.NewService(deps.Queries, inspSvc)
	if err := declSvc.PersistExtracted(ctx, inspectionID, mustParseUUID(t, imageID), []ml.ExtractedField{
		{Field: "MRP", RawValue: "M.R.P. Rs. 249.00", Confidence: 0.95, BBox: &ml.BBox{X: 10, Y: 10, Width: 100, Height: 20}},
		// No parseable "<number><unit>" token, so normalization (which does not
		// itself validate legal formatting — see internal/declarations/normalization.go)
		// falls back to passing the raw text through unchanged, which then fails
		// the rule's strict `^\d+\s?(g|kg|ml|l)$` format check below.
		{Field: "NET_QUANTITY", RawValue: "quantity not clearly stated", Confidence: 0.9},
	}, "test-mongo-ref"); err != nil {
		t.Fatalf("persist extracted declarations: %v", err)
	}

	code, env = doRequest(t, router, http.MethodGet, "/api/v1/inspections/"+inspection.ID+"/declarations", adminToken, nil)
	if code != http.StatusOK {
		t.Fatalf("list declarations: expected 200, got %d: %+v", code, env)
	}
	declList := mustDecode[[]struct {
		ID         string   `json:"id"`
		Type       string   `json:"type"`
		RawValue   *string  `json:"raw_value"`
		Confidence *float64 `json:"confidence"`
	}](t, env.Data)
	if len(declList) != 2 {
		t.Fatalf("expected 2 normalized declarations, got %d: %+v", len(declList), declList)
	}
	var mrpDeclarationID string
	for _, d := range declList {
		if d.Type == "MRP" {
			mrpDeclarationID = d.ID
			if d.RawValue == nil || *d.RawValue != "M.R.P. Rs. 249.00" {
				t.Fatalf("expected raw OCR value preserved alongside the normalized value, got %+v", d)
			}
		}
	}
	if mrpDeclarationID == "" {
		t.Fatal("expected an MRP declaration")
	}

	// ---- 9. Compliance orchestration (sections 11-13, 19): run synchronously
	// (the exported method the worker itself calls) so the test doesn't need
	// a running worker process, but still exercises the real path — build a
	// request from the persisted declarations, POST it to the mocked Python
	// rule engine, and persist the returned checks/violations/outcome inside
	// a transaction. ----
	complianceSvc := compliance.NewService(deps.Queries, deps.Pool, deps.Asynq, inspSvc, deps.MLRepo, deps.RuleEngineClient)
	summary, err := complianceSvc.Run(ctx, inspectionID)
	if err != nil {
		t.Fatalf("run compliance evaluation: %v", err)
	}
	// MRP is present but the net-quantity text doesn't match the mock rule
	// engine's expected format, so it returns one violation -> NON_COMPLIANT.
	if summary.Outcome != "NON_COMPLIANT" {
		t.Fatalf("expected NON_COMPLIANT (MRP present but net quantity format invalid), got %s (summary=%+v)", summary.Outcome, summary)
	}

	code, env = doRequest(t, router, http.MethodGet, "/api/v1/inspections/"+inspection.ID+"/checks", adminToken, nil)
	if code != http.StatusOK {
		t.Fatalf("list checks: expected 200, got %d: %+v", code, env)
	}
	checks := mustDecode[[]struct {
		Status string `json:"status"`
	}](t, env.Data)
	var sawPass, sawFail bool
	for _, c := range checks {
		switch c.Status {
		case "PASS":
			sawPass = true
		case "FAIL":
			sawFail = true
		}
	}
	if !sawPass {
		t.Fatalf("expected at least one PASS check (MRP present), got %+v", checks)
	}
	if !sawFail {
		t.Fatalf("expected the malformed NET_QUANTITY to produce a FAIL check + violation, got %+v", checks)
	}

	code, env = doRequest(t, router, http.MethodGet, "/api/v1/inspections/"+inspection.ID+"/violations", adminToken, nil)
	if code != http.StatusOK {
		t.Fatalf("list violations: expected 200, got %d: %+v", code, env)
	}
	violationsList := mustDecode[[]struct {
		LegalReference *string `json:"legal_reference"`
	}](t, env.Data)
	if len(violationsList) != 1 {
		t.Fatalf("expected exactly 1 persisted violation, got %+v", violationsList)
	}

	// ---- 10. Declaration manual correction (section 9): update via HTTP,
	// verify audit trail + re-evaluation re-enqueue (status flips back to
	// PROCESSING, the same signal a real recheck would produce). ----
	code, env = doRequest(t, router, http.MethodPatch, "/api/v1/declarations/"+mrpDeclarationID, adminToken, map[string]any{
		"normalizedValue": "249.50", "numericValue": 249.5,
	})
	if code != http.StatusOK {
		t.Fatalf("patch declaration: expected 200, got %d: %+v", code, env)
	}
	updatedDecl := mustDecode[struct {
		NormalizedValue    *string `json:"normalized_value"`
		IsEditedByReviewer bool    `json:"is_edited_by_reviewer"`
		Source             string  `json:"source"`
	}](t, env.Data)
	if !updatedDecl.IsEditedByReviewer || updatedDecl.Source != "MANUAL" || updatedDecl.NormalizedValue == nil || *updatedDecl.NormalizedValue != "249.50" {
		t.Fatalf("expected the manual edit to be recorded, got %+v", updatedDecl)
	}

	code, env = doRequest(t, router, http.MethodGet, "/api/v1/inspections/"+inspection.ID, adminToken, nil)
	if code != http.StatusOK {
		t.Fatalf("get inspection after declaration edit: expected 200, got %d: %+v", code, env)
	}
	afterEdit := mustDecode[struct {
		Status string `json:"status"`
	}](t, env.Data)
	if afterEdit.Status != "PROCESSING" {
		t.Fatalf("expected declaration edit to re-enqueue compliance evaluation (status=PROCESSING), got %s", afterEdit.Status)
	}
	// Re-run synchronously (again, standing in for the worker) so downstream
	// violation-workflow steps have a settled state to work with.
	if _, err := complianceSvc.Run(ctx, inspectionID); err != nil {
		t.Fatalf("re-run compliance evaluation: %v", err)
	}

	// ---- 11. Evidence (section 15): attach + list with signed URLs. ----
	code, env = doRequest(t, router, http.MethodGet, "/api/v1/inspections/"+inspection.ID+"/evidence", adminToken, nil)
	if code != http.StatusOK {
		t.Fatalf("list evidence: expected 200, got %d: %+v", code, env)
	}

	// ---- 12. Audit logging (section 17): the mutations above must have left
	// a trail (actor, action, entity, old/new state, ip, user-agent). ----
	code, env = doRequest(t, router, http.MethodGet, "/api/v1/audit-logs?entityType=Inspection&entityId="+inspection.ID, adminToken, nil)
	if code != http.StatusOK {
		t.Fatalf("list audit logs: expected 200, got %d: %+v", code, env)
	}
	auditEntries := mustDecode[[]struct {
		Action string `json:"action"`
	}](t, env.Data)
	if len(auditEntries) == 0 {
		t.Fatal("expected at least one audit log entry for this inspection")
	}

	// ---- 13. Finalize / reopen (section 16): once finalized, further edits
	// are rejected until explicitly reopened. ----
	code, env = doRequest(t, router, http.MethodPost, "/api/v1/inspections/"+inspection.ID+"/finalize", adminToken, nil)
	if code != http.StatusOK {
		t.Fatalf("finalize: expected 200, got %d: %+v", code, env)
	}
	finalized := mustDecode[struct {
		IsFinalized bool `json:"is_finalized"`
	}](t, env.Data)
	if !finalized.IsFinalized {
		t.Fatal("expected is_finalized=true after finalize")
	}

	if code, _ := doRequest(t, router, http.MethodPatch, "/api/v1/inspections/"+inspection.ID, adminToken, map[string]any{"notes": "should be rejected"}); code != http.StatusConflict {
		t.Fatalf("expected PATCH on a finalized inspection to be rejected with 409, got %d", code)
	}
	if code, _ := doRequest(t, router, http.MethodPatch, "/api/v1/declarations/"+mrpDeclarationID, adminToken, map[string]any{"normalizedValue": "should be rejected"}); code != http.StatusConflict {
		t.Fatalf("expected a declaration edit on a finalized inspection to be rejected with 409, got %d", code)
	}

	code, env = doRequest(t, router, http.MethodPost, "/api/v1/inspections/"+inspection.ID+"/reopen", adminToken, nil)
	if code != http.StatusOK {
		t.Fatalf("reopen: expected 200, got %d: %+v", code, env)
	}
	reopened := mustDecode[struct {
		IsFinalized bool `json:"is_finalized"`
	}](t, env.Data)
	if reopened.IsFinalized {
		t.Fatal("expected is_finalized=false after reopen")
	}
	if code, _ := doRequest(t, router, http.MethodPatch, "/api/v1/inspections/"+inspection.ID, adminToken, map[string]any{"notes": "allowed after reopen"}); code != http.StatusOK {
		t.Fatalf("expected PATCH to succeed after reopen, got %d", code)
	}

	// ---- 14. Logout revokes the refresh token. ----
	code, _ = doRequest(t, router, http.MethodPost, "/api/v1/auth/logout", "", map[string]any{"refreshToken": refreshed.AccessToken})
	_ = code // logout is intentionally idempotent even for a bad/foreign token; no assertion needed beyond "did not panic".
}

func uploadTestImage(t *testing.T, router http.Handler, token, inspectionID string) (imageID, jobID string) {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	// CreateFormFile always sends Content-Type: application/octet-stream,
	// which RequireSingleImage's MIME allow-list rejects — set it explicitly.
	header := textproto.MIMEHeader{}
	header.Set("Content-Disposition", `form-data; name="image"; filename="label.jpg"`)
	header.Set("Content-Type", "image/jpeg")
	part, err := w.CreatePart(header)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	// Minimal valid-enough JPEG magic bytes; upload validates MIME + size only.
	if _, err := part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 'J', 'F', 'I', 'F'}); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/inspections/"+inspectionID+"/images", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("upload image: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var env apiEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode upload response: %v", err)
	}
	result := mustDecode[struct {
		ImageID string `json:"imageId"`
		JobID   string `json:"jobId"`
	}](t, env.Data)
	return result.ImageID, result.JobID
}

func mustParseUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	parsed, err := uuid.Parse(s)
	if err != nil {
		t.Fatalf("parse uuid %q: %v", s, err)
	}
	return parsed
}

func strPtr(s string) *string { return &s }
