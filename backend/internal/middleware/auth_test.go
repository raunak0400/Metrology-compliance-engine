package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/authtoken"
)

func init() { gin.SetMode(gin.TestMode) }

func testTokenizer() *authtoken.Tokenizer {
	return authtoken.NewTokenizer("access-secret-at-least-8", "refresh-secret-at-least-8", 15*time.Minute, 7*24*time.Hour)
}

// withAsUser injects a fixed AuthUser into the context, standing in for
// Authenticate in tests that only care about Authorize/RequirePermission.
func withAsUser(u AuthUser) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(userContextKey, u)
		c.Next()
	}
}

// newProtectedRouter wires the chain the real app uses: ErrorHandler wraps
// everything (registered first, so its post-c.Next() error handling still
// runs even if a later handler calls c.Abort()), matching cmd/api/main.go's
// v1.Use(ErrorHandler) before any route-group's g.Use(Authenticate).
func newProtectedRouter(routeMiddleware ...gin.HandlerFunc) *gin.Engine {
	router := gin.New()
	router.Use(ErrorHandler(false))
	router.GET("/protected", append(routeMiddleware, func(c *gin.Context) { c.Status(http.StatusOK) })...)
	return router
}

func doGet(router *gin.Engine, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAuthenticate_MissingToken_Returns401ViaErrorChain(t *testing.T) {
	router := newProtectedRouter(Authenticate(testTokenizer()))
	rec := doGet(router, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing bearer token, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAuthenticate_InvalidToken_Returns401(t *testing.T) {
	router := newProtectedRouter(Authenticate(testTokenizer()))
	rec := doGet(router, "not-a-real-jwt")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for an invalid token, got %d", rec.Code)
	}
}

func TestAuthenticate_ValidToken_PopulatesCurrentUser(t *testing.T) {
	tok := testTokenizer()
	access, err := tok.SignAccessToken("user-1", "inspector@example.test", []string{"INSPECTOR"}, []string{"inspections:write"})
	if err != nil {
		t.Fatalf("sign access token: %v", err)
	}

	var captured AuthUser
	router := gin.New()
	router.Use(ErrorHandler(false), Authenticate(tok))
	router.GET("/protected", func(c *gin.Context) {
		captured, _ = CurrentUser(c)
		c.Status(http.StatusOK)
	})

	rec := doGet(router, access)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for a valid token, got %d: %s", rec.Code, rec.Body.String())
	}
	if captured.ID != "user-1" || !captured.HasRole("INSPECTOR") {
		t.Fatalf("expected AuthUser to be populated from claims, got %+v", captured)
	}
}

func TestAuthorize_AdminAlwaysPasses_RegardlessOfListedRoles(t *testing.T) {
	router := newProtectedRouter(withAsUser(AuthUser{ID: "u1", Roles: []string{"ADMIN"}}), Authorize("SUPERVISOR"))
	rec := doGet(router, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected ADMIN to bypass the role list, got %d", rec.Code)
	}
}

func TestAuthorize_WrongRole_Returns403(t *testing.T) {
	router := newProtectedRouter(withAsUser(AuthUser{ID: "u1", Roles: []string{"INSPECTOR"}}), Authorize("SUPERVISOR", "REVIEWER"))
	rec := doGet(router, "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for a role not in the allow-list, got %d", rec.Code)
	}
}

func TestAuthorize_ListedRole_Passes(t *testing.T) {
	router := newProtectedRouter(withAsUser(AuthUser{ID: "u1", Roles: []string{"REVIEWER"}}), Authorize("SUPERVISOR", "REVIEWER"))
	rec := doGet(router, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected a listed role to pass, got %d", rec.Code)
	}
}

func TestRequirePermission_WildcardGrantsEverything(t *testing.T) {
	router := newProtectedRouter(withAsUser(AuthUser{ID: "u1", Permissions: []string{"*"}}), RequirePermission("inspections:write", "rules:write"))
	rec := doGet(router, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected wildcard permission to grant access, got %d", rec.Code)
	}
}

func TestRequirePermission_MissingScope_Returns403(t *testing.T) {
	router := newProtectedRouter(withAsUser(AuthUser{ID: "u1", Permissions: []string{"inspections:read"}}), RequirePermission("inspections:write"))
	rec := doGet(router, "")
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 when the required permission scope is missing, got %d", rec.Code)
	}
}

func TestRequirePermission_HasAllScopes_Passes(t *testing.T) {
	router := newProtectedRouter(
		withAsUser(AuthUser{ID: "u1", Permissions: []string{"inspections:read", "inspections:write"}}),
		RequirePermission("inspections:read", "inspections:write"),
	)
	rec := doGet(router, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected access when all required scopes are held, got %d", rec.Code)
	}
}

func TestOptionalAuth_NoToken_DoesNotAbort(t *testing.T) {
	router := newProtectedRouter(OptionalAuth(testTokenizer()))
	rec := doGet(router, "")
	if rec.Code != http.StatusOK {
		t.Fatalf("expected OptionalAuth to allow anonymous requests through, got %d", rec.Code)
	}
}
