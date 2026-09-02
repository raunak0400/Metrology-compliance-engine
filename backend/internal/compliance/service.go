// Compliance orchestration.
//
// Loads structured declarations + measurements from PostgreSQL and the
// latest raw ML analysis from MongoDB, builds a request for the Python
// Legal Metrology Rule Engine service, and persists whatever it returns —
// compliance checks, violations, and the inspection outcome. The Go backend
// never decides compliance itself; only the external rule engine does.
//
// Pipeline position:
//
//	declarations + measurements + package metadata -> POST /validate
//	  -> ValidationResponse -> checks + violations (PostgreSQL)
//	                        -> raw request/response (MongoDB evidence trail)
package compliance

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/raunak0400/metrology-backend/internal/db"
	"github.com/raunak0400/metrology-backend/internal/dbutil"
	"github.com/raunak0400/metrology-backend/internal/inspections"
	"github.com/raunak0400/metrology-backend/internal/ml"
	"github.com/raunak0400/metrology-backend/internal/ruleengine"
	"github.com/raunak0400/metrology-backend/internal/taskqueue"
)

type Service struct {
	q          *db.Queries
	pool       *pgxpool.Pool
	asynq      *asynq.Client
	inspSvc    *inspections.Service
	mlRepo     *ml.Repository
	ruleClient *ruleengine.Client
}

func NewService(q *db.Queries, pool *pgxpool.Pool, asynqClient *asynq.Client, inspSvc *inspections.Service, mlRepo *ml.Repository, ruleClient *ruleengine.Client) *Service {
	return &Service{q: q, pool: pool, asynq: asynqClient, inspSvc: inspSvc, mlRepo: mlRepo, ruleClient: ruleClient}
}

type RunSummary struct {
	InspectionID string               `json:"inspectionId"`
	Outcome      db.InspectionOutcome `json:"outcome"`
	TotalChecks  int                  `json:"totalChecks"`
	Passed       int                  `json:"passed"`
	Failed       int                  `json:"failed"`
	ManualReview int                  `json:"manualReview"`
	Violations   int                  `json:"violations"`
}

// Run calls the rule engine service and persists the results. The rule
// engine is the sole decision-maker; this method only assembles its input
// and stores its output.
func (s *Service) Run(ctx context.Context, inspectionID uuid.UUID) (RunSummary, error) {
	insp, err := s.inspSvc.Ensure(ctx, inspectionID)
	if err != nil {
		return RunSummary{}, err
	}

	declRows, err := s.q.ListDeclarationsByInspection(ctx, inspectionID)
	if err != nil {
		return RunSummary{}, err
	}

	scaleRows, err := s.q.ListDeclarationsWithScaleByInspection(ctx, inspectionID)
	if err != nil {
		return RunSummary{}, err
	}
	scaleInput := make([]DeclarationScaleRow, 0, len(scaleRows))
	for _, r := range scaleRows {
		scaleInput = append(scaleInput, DeclarationScaleRow{DeclarationID: r.ID.String(), Type: string(r.Type), BBoxJSON: r.Bbox, ScaleReference: r.ScaleReference})
	}
	heights := DeclarationHeightsMM(scaleInput)

	req := s.buildValidationRequest(ctx, insp, declRows, heights)

	resp, err := s.ruleClient.Validate(ctx, req)
	if err != nil {
		return RunSummary{}, err
	}

	if _, err := s.mlRepo.SaveMLResult(ctx, "RULE_ENGINE_VALIDATION", inspectionID.String(), "", resp.RulesetVersion, map[string]interface{}{
		"request": req, "response": resp,
	}); err != nil {
		return RunSummary{}, err
	}

	ruleIDs, err := s.syncRuleRegistry(ctx)
	if err != nil {
		return RunSummary{}, err
	}

	if err := s.persistResults(ctx, inspectionID, *resp, ruleIDs); err != nil {
		return RunSummary{}, err
	}

	violated := make(map[string]bool, len(resp.Violations))
	for _, v := range resp.Violations {
		violated[v.RuleID] = true
	}
	summary := RunSummary{InspectionID: inspectionID.String(), Outcome: mapOutcome(resp.OverallStatus), Failed: len(resp.Violations), Violations: len(resp.Violations)}
	for _, note := range resp.ApplicabilityNotes {
		if violated[note.RuleID] {
			continue
		}
		if note.Applied {
			summary.Passed++
		}
	}
	summary.ManualReview = len(resp.ManualReviewRequired)
	summary.TotalChecks = len(resp.ApplicabilityNotes)
	return summary, nil
}

// declarationTypeToFieldType is a best-effort, non-authoritative shape
// translation from this backend's DeclarationType enum to the rule engine's
// field_type vocabulary. It is NOT a legal decision — the rule engine passes
// unrecognized field_type values through unvalidated (see rule_engine/schemas.py).
var declarationTypeToFieldType = map[string]string{
	"MRP":               "mrp",
	"NET_QUANTITY":      "net_quantity",
	"MANUFACTURER":      "manufacturer",
	"PACKER":            "manufacturer",
	"IMPORTER":          "importer",
	"MANUFACTURE_DATE":  "mfg_date",
	"CONSUMER_CARE":     "consumer_care",
	"COUNTRY_OF_ORIGIN": "country_of_origin",
	"GENERIC_NAME":      "generic_name",
}

func toFieldType(declType string) string {
	if v, ok := declarationTypeToFieldType[declType]; ok {
		return v
	}
	return strings.ToLower(declType)
}

func normalizeEnum(v *string, allowed map[string]bool, def string) string {
	if v == nil {
		return def
	}
	lc := strings.ToLower(strings.TrimSpace(*v))
	if allowed[lc] {
		return lc
	}
	return def
}

var allowedCategories = map[string]bool{"food": true, "cosmetics": true, "electronics": true, "other": true, "cement": true, "fertilizer": true, "agricultural_produce": true}

func normalizePackageShape(v *string) string {
	if v == nil {
		return "irregular"
	}
	lc := strings.ToLower(*v)
	switch {
	case strings.Contains(lc, "cylind"):
		return "cylindrical"
	case strings.Contains(lc, "rect") || strings.Contains(lc, "box") || strings.Contains(lc, "cuboid"):
		return "rectangular"
	default:
		return "irregular"
	}
}

func normalizeSurfaceType(v *string) string {
	if v == nil {
		return "printed"
	}
	lc := strings.ToLower(*v)
	if strings.Contains(lc, "emboss") || strings.Contains(lc, "mold") || strings.Contains(lc, "blow") {
		return "embossed_molded_blown"
	}
	return "printed"
}

func normalizeSaleChannel(v *string) string {
	if v == nil {
		return "retail"
	}
	lc := strings.ToLower(*v)
	if lc == "ecommerce" || lc == "e-commerce" || lc == "online" {
		return "ecommerce"
	}
	return "retail"
}

type declBBox struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// buildValidationRequest assembles the /validate request body: package-level
// metadata comes from the latest raw ML analysis (MongoDB, best-effort — a
// missing document just falls back to sane defaults), the field list comes
// from the authoritative, possibly-reviewer-edited PostgreSQL declarations.
func (s *Service) buildValidationRequest(ctx context.Context, insp db.Inspection, declarations []db.ExtractedDeclaration, heights map[string]*float64) ruleengine.ValidationRequest {
	var analysis ml.AnalyzeResult
	_ = s.mlRepo.GetLatestMLResult(ctx, "ANALYZE", insp.ID.String(), &analysis)

	productID := insp.ID.String()
	if analysis.ProductID != nil && *analysis.ProductID != "" {
		productID = *analysis.ProductID
	}

	fields := make([]ruleengine.ExtractedField, 0, len(declarations))
	measurements := make([]ruleengine.TextMeasurementEntry, 0, len(declarations))
	isImported := false

	for _, d := range declarations {
		if d.Type == db.DeclarationTypeIMPORTER && d.NormalizedValue != nil && strings.TrimSpace(*d.NormalizedValue) != "" {
			isImported = true
		}

		text := dbutil.StrVal(d.RawValue)
		if text == "" {
			text = dbutil.StrVal(d.NormalizedValue)
		}
		confidence := 0.0
		if d.Confidence != nil {
			confidence = *d.Confidence
		}

		var bbox ruleengine.BoundingBox
		if len(d.Bbox) > 0 {
			var b declBBox
			if err := json.Unmarshal(d.Bbox, &b); err == nil {
				bbox = ruleengine.BoundingBox{X: b.X, Y: b.Y, W: b.Width, H: b.Height}
			}
		}

		fontHeight := heights[d.ID.String()]

		fields = append(fields, ruleengine.ExtractedField{
			FieldType:     toFieldType(string(d.Type)),
			Text:          text,
			Confidence:    confidence,
			BBox:          bbox,
			FontHeightMM:  fontHeight,
			PanelLocation: "principal_display_panel",
		})
		if fontHeight != nil {
			measurements = append(measurements, ruleengine.TextMeasurementEntry{
				Text: text, Confidence: confidence, HeightMM: *fontHeight, MinRequiredMM: 0, IsRule7Compliant: true,
			})
		}
	}

	if analysis.IsImported != nil {
		isImported = isImported || *analysis.IsImported
	}

	return ruleengine.ValidationRequest{
		ProductID:          productID,
		Category:           normalizeEnum(analysis.Category, allowedCategories, "other"),
		PackageShape:       normalizePackageShape(analysis.PackageShape),
		PackageSurfaceType: normalizeSurfaceType(analysis.PackageSurfaceType),
		SaleChannel:        normalizeSaleChannel(analysis.SaleChannel),
		NetWeightKg:        analysis.NetWeightKg,
		IsImported:         isImported,
		ExtractedFields:    fields,
		TextMeasurements:   measurements,
	}
}

func mapSeverity(s string) db.ViolationSeverity {
	switch strings.ToLower(s) {
	case "major":
		return db.ViolationSeverityHIGH
	case "minor":
		return db.ViolationSeverityMEDIUM
	case "advisory":
		return db.ViolationSeverityLOW
	default:
		return db.ViolationSeverityMEDIUM
	}
}

func mapOutcome(status string) db.InspectionOutcome {
	switch status {
	case "compliant", "exempt":
		return db.InspectionOutcomeCOMPLIANT
	case "non_compliant":
		return db.InspectionOutcomeNONCOMPLIANT
	default: // requires_manual_review, or anything unrecognized — never silently COMPLIANT
		return db.InspectionOutcomeMANUALREVIEW
	}
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}

// syncRuleRegistry mirrors the Python rule engine's rule registry (GET
// /rules) into the local legal_rules table by code, and returns a map from
// the engine's rule_id to the local row's UUID — needed only because
// compliance_checks.rule_id is a NOT NULL foreign key. It never drives
// evaluation; the mirrored rows carry no requirements.
func (s *Service) syncRuleRegistry(ctx context.Context) (map[string]uuid.UUID, error) {
	registry, err := s.ruleClient.Rules(ctx)
	if err != nil {
		return nil, err
	}
	ids := make(map[string]uuid.UUID, len(registry.Rules))
	for _, r := range registry.Rules {
		row, err := s.q.UpsertLegalRuleByCode(ctx, db.UpsertLegalRuleByCodeParams{
			Code: r.RuleID, Title: r.Description,
			Description: strPtr(r.Description), ActReference: strPtr(r.Clause),
			SeverityDefault: mapSeverity(r.Severity),
		})
		if err != nil {
			return nil, err
		}
		ids[r.RuleID] = row.ID
	}
	return ids, nil
}

// persistResults persists checks + violations and updates the inspection
// outcome inside a single PostgreSQL transaction — either the full evaluation
// lands, or none of it does. Idempotent re-run: clears the prior evaluation
// first.
func (s *Service) persistResults(ctx context.Context, inspectionID uuid.UUID, resp ruleengine.ValidationResponse, ruleIDs map[string]uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := s.q.WithTx(tx)

	if err := q.DeleteViolationsByInspection(ctx, inspectionID); err != nil {
		return err
	}
	if err := q.DeleteComplianceChecksByInspection(ctx, inspectionID); err != nil {
		return err
	}

	handled := make(map[string]bool, len(resp.Violations))

	for _, v := range resp.Violations {
		ruleUUID, ok := ruleIDs[v.RuleID]
		if !ok {
			continue // rule engine referenced a rule_id its own /rules registry didn't list — skip rather than fail the whole run
		}
		handled[v.RuleID] = true

		reason := strPtr(v.Description)
		details := dbutil.JSONBytes(map[string]interface{}{"ruleId": v.RuleID, "clause": v.Clause, "fieldType": v.FieldType})

		check, err := q.CreateComplianceCheck(ctx, db.CreateComplianceCheckParams{
			InspectionID: inspectionID, RuleID: ruleUUID, Status: db.ComplianceStatusFAIL,
			ActualValue: v.FieldEvidence, Reason: reason, Details: details,
		})
		if err != nil {
			return err
		}

		if _, err := q.CreateViolation(ctx, db.CreateViolationParams{
			InspectionID: inspectionID, RuleID: uuid.NullUUID{UUID: ruleUUID, Valid: true},
			CheckID: uuid.NullUUID{UUID: check.ID, Valid: true}, Severity: mapSeverity(v.Severity),
			Title: truncate(v.Description, 250), Reason: reason, LegalReference: strPtr(v.Clause),
		}); err != nil {
			return err
		}
	}

	for _, note := range resp.ApplicabilityNotes {
		if handled[note.RuleID] {
			continue
		}
		ruleUUID, ok := ruleIDs[note.RuleID]
		if !ok {
			continue
		}
		status := db.ComplianceStatusPASS
		if !note.Applied {
			status = db.ComplianceStatusNOTAPPLICABLE
		}
		details := dbutil.JSONBytes(map[string]interface{}{"ruleId": note.RuleID})
		if _, err := q.CreateComplianceCheck(ctx, db.CreateComplianceCheckParams{
			InspectionID: inspectionID, RuleID: ruleUUID, Status: status, Reason: strPtr(note.Reason), Details: details,
		}); err != nil {
			return err
		}
	}

	outcome := mapOutcome(resp.OverallStatus)
	status := db.InspectionStatusCOMPLETED
	if outcome == db.InspectionOutcomeMANUALREVIEW {
		status = db.InspectionStatusAWAITINGREVIEW
	}
	if err := q.SetInspectionOutcome(ctx, db.SetInspectionOutcomeParams{ID: inspectionID, Status: status, Outcome: outcome}); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// Enqueue schedules a compliance evaluation on the compliance-check queue.
func (s *Service) Enqueue(ctx context.Context, inspectionID uuid.UUID, triggeredByUserID string) (string, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return "", err
	}

	payload, err := json.Marshal(taskqueue.CompliancePayload{InspectionID: inspectionID.String(), TriggeredByUserID: triggeredByUserID})
	if err != nil {
		return "", err
	}
	info, err := s.asynq.Enqueue(asynq.NewTask(taskqueue.TaskComplianceCheck, payload), asynq.Queue(taskqueue.QueueCompliance))
	if err != nil {
		return "", err
	}
	_, _ = s.q.UpdateInspection(ctx, db.UpdateInspectionParams{ID: inspectionID, Status: statusPtr(db.InspectionStatusPROCESSING)})
	return info.ID, nil
}

func statusPtr(s db.InspectionStatus) *db.InspectionStatus { return &s }

func (s *Service) ListChecks(ctx context.Context, inspectionID uuid.UUID) ([]db.ListComplianceChecksByInspectionRow, error) {
	if _, err := s.inspSvc.Ensure(ctx, inspectionID); err != nil {
		return nil, err
	}
	return s.q.ListComplianceChecksByInspection(ctx, inspectionID)
}
