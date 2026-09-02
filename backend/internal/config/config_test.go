package config

import (
	"strings"
	"testing"
	"time"
)

// clearEnv removes every env var Load() reads, so each test starts from a
// clean slate regardless of what's set in the host shell / CI runner.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"GO_ENV", "PORT", "LOG_LEVEL", "API_PREFIX",
		"DATABASE_URL", "MONGODB_URI", "REDIS_URL",
		"JWT_ACCESS_SECRET", "JWT_REFRESH_SECRET", "JWT_ACCESS_TTL", "JWT_REFRESH_TTL",
		"S3_ENDPOINT", "S3_REGION", "S3_ACCESS_KEY", "S3_SECRET_KEY", "S3_BUCKET",
		"S3_FORCE_PATH_STYLE", "S3_USE_SSL", "S3_PUBLIC_URL", "S3_SIGNED_URL_TTL",
		"ML_SERVICE_URL", "ML_SERVICE_TIMEOUT_MS", "ML_SERVICE_API_KEY",
		"MAX_UPLOAD_SIZE_BYTES", "ALLOWED_IMAGE_MIME",
		"RATE_LIMIT_WINDOW_MS", "RATE_LIMIT_MAX", "CORS_ORIGINS",
	} {
		// Load() treats an empty string identically to unset (os.Getenv
		// returns "" for both); t.Setenv also restores the pre-test value
		// automatically on cleanup, which plain os.Unsetenv would not.
		t.Setenv(key, "")
	}
}

func validRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/legal_metrology?sslmode=disable")
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017/legal_metrology_ml")
	t.Setenv("REDIS_URL", "redis://localhost:6379")
	t.Setenv("JWT_ACCESS_SECRET", "access-secret-long-enough")
	t.Setenv("JWT_REFRESH_SECRET", "refresh-secret-long-enough")
	t.Setenv("S3_ENDPOINT", "localhost:9000")
	t.Setenv("S3_ACCESS_KEY", "minioadmin")
	t.Setenv("S3_SECRET_KEY", "minioadmin")
	t.Setenv("S3_BUCKET", "legal-metrology")
	t.Setenv("ML_SERVICE_URL", "http://localhost:8000")
}

func TestLoad_MissingRequiredVars_FailsWithAllProblemsListed(t *testing.T) {
	clearEnv(t)

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error when required env vars are missing")
	}

	for _, want := range []string{"DATABASE_URL", "MONGODB_URI", "REDIS_URL", "JWT_ACCESS_SECRET", "JWT_REFRESH_SECRET", "S3_ENDPOINT", "S3_ACCESS_KEY", "S3_SECRET_KEY", "S3_BUCKET", "ML_SERVICE_URL"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected error to mention missing %s, got: %s", want, err.Error())
		}
	}
}

func TestLoad_ValidEnv_Succeeds(t *testing.T) {
	clearEnv(t)
	validRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.GoEnv != "development" {
		t.Errorf("expected default GoEnv=development, got %s", cfg.GoEnv)
	}
	if cfg.Port != 4000 {
		t.Errorf("expected default Port=4000, got %d", cfg.Port)
	}
	if cfg.JWTAccessTTL != 15*time.Minute {
		t.Errorf("expected default JWTAccessTTL=15m, got %s", cfg.JWTAccessTTL)
	}
	if cfg.JWTRefreshTTL != 7*24*time.Hour {
		t.Errorf("expected default JWTRefreshTTL=7d, got %s", cfg.JWTRefreshTTL)
	}
	if !cfg.S3ForcePathStyle {
		t.Error("expected default S3ForcePathStyle=true")
	}
}

func TestLoad_InvalidGoEnv_Fails(t *testing.T) {
	clearEnv(t)
	validRequiredEnv(t)
	t.Setenv("GO_ENV", "not-a-real-env")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for an invalid GO_ENV value")
	}
	if !strings.Contains(err.Error(), "GO_ENV") {
		t.Errorf("expected error to mention GO_ENV, got: %s", err.Error())
	}
}

func TestLoad_ShortJWTSecret_Fails(t *testing.T) {
	clearEnv(t)
	validRequiredEnv(t)
	t.Setenv("JWT_ACCESS_SECRET", "short")

	_, err := Load()
	if err == nil {
		t.Fatal("expected an error for a JWT secret under 8 characters")
	}
}

func TestLoad_DurationShorthand_ParsesDaySuffix(t *testing.T) {
	clearEnv(t)
	validRequiredEnv(t)
	t.Setenv("JWT_REFRESH_TTL", "14d")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.JWTRefreshTTL != 14*24*time.Hour {
		t.Errorf("expected 14d to parse as 336h, got %s", cfg.JWTRefreshTTL)
	}
}
