package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/raunak0400/metrology-backend/internal/app"
	"github.com/raunak0400/metrology-backend/internal/config"
)

// TestHealthEndpoint verifies the app boots and /health responds without
// needing any live infrastructure (Postgres/Mongo/Redis) — mirroring the
// original Node service's opt-in-by-default integration test philosophy,
// but kept infra-free since /health never touches deps.
func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testConfig()
	router := buildRouter(cfg, &app.Deps{Config: cfg})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestNotFoundEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testConfig()
	router := buildRouter(cfg, &app.Deps{Config: cfg})

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func testConfig() *config.Config {
	return &config.Config{
		APIPrefix:          "/api/v1",
		CORSOrigins:        []string{"http://localhost:3000"},
		RateLimitWindow:    0,
		RateLimitMax:       1000,
		MaxUploadSizeBytes: 15_728_640,
		AllowedImageMIME:   []string{"image/jpeg", "image/png", "image/webp"},
	}
}
