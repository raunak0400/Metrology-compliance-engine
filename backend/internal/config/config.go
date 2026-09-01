// Package config loads and validates process configuration from the
// environment. Import this instead of reading os.Getenv directly anywhere
// else in the codebase — mirrors the fail-fast behavior of the original
// Node service's src/config/env.ts.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	GoEnv     string // development | test | production
	Port      int
	LogLevel  string
	APIPrefix string

	DatabaseURL string
	MongoDBURI  string
	RedisURL    string

	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration

	S3Endpoint       string
	S3Region         string
	S3AccessKey      string
	S3SecretKey      string
	S3Bucket         string
	S3ForcePathStyle bool
	S3UseSSL         bool
	S3PublicURL      string
	S3SignedURLTTL   time.Duration

	MLServiceURL     string
	MLServiceTimeout time.Duration
	MLServiceAPIKey  string

	MaxUploadSizeBytes int64
	AllowedImageMIME   []string

	RateLimitWindow time.Duration
	RateLimitMax    int

	CORSOrigins []string
}

func (c Config) IsProd() bool { return c.GoEnv == "production" }
func (c Config) IsTest() bool { return c.GoEnv == "test" }

// Load reads and validates configuration from the environment. It returns an
// error listing every missing/invalid variable rather than failing on the
// first one, matching the readable "fail fast" behavior of the Node service.
func Load() (*Config, error) {
	var problems []string

	req := func(key string) string {
		v := os.Getenv(key)
		if v == "" {
			problems = append(problems, fmt.Sprintf("%s is required", key))
		}
		return v
	}
	optDefault := func(key, def string) string {
		v := os.Getenv(key)
		if v == "" {
			return def
		}
		return v
	}
	optInt := func(key string, def int) int {
		v := os.Getenv(key)
		if v == "" {
			return def
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s must be an integer", key))
			return def
		}
		return n
	}
	optInt64 := func(key string, def int64) int64 {
		v := os.Getenv(key)
		if v == "" {
			return def
		}
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s must be an integer", key))
			return def
		}
		return n
	}
	optBool := func(key string, def bool) bool {
		v := os.Getenv(key)
		if v == "" {
			return def
		}
		return v == "true" || v == "1"
	}
	optDuration := func(key string, def time.Duration) time.Duration {
		v := os.Getenv(key)
		if v == "" {
			return def
		}
		d, err := parseGoOrJSDuration(v)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s must be a duration (e.g. 15m, 7d)", key))
			return def
		}
		return d
	}
	csv := func(key, def string) []string {
		v := optDefault(key, def)
		parts := strings.Split(v, ",")
		out := make([]string, 0, len(parts))
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				out = append(out, p)
			}
		}
		return out
	}

	cfg := &Config{
		GoEnv:     optDefault("GO_ENV", "development"),
		Port:      optInt("PORT", 4000),
		LogLevel:  optDefault("LOG_LEVEL", "info"),
		APIPrefix: optDefault("API_PREFIX", "/api/v1"),

		DatabaseURL: req("DATABASE_URL"),
		MongoDBURI:  req("MONGODB_URI"),
		RedisURL:    req("REDIS_URL"),

		JWTAccessSecret:  req("JWT_ACCESS_SECRET"),
		JWTRefreshSecret: req("JWT_REFRESH_SECRET"),
		JWTAccessTTL:     optDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:    optDuration("JWT_REFRESH_TTL", 7*24*time.Hour),

		S3Endpoint:       req("S3_ENDPOINT"),
		S3Region:         optDefault("S3_REGION", "us-east-1"),
		S3AccessKey:      req("S3_ACCESS_KEY"),
		S3SecretKey:      req("S3_SECRET_KEY"),
		S3Bucket:         req("S3_BUCKET"),
		S3ForcePathStyle: optBool("S3_FORCE_PATH_STYLE", true),
		S3UseSSL:         optBool("S3_USE_SSL", false),
		S3PublicURL:      os.Getenv("S3_PUBLIC_URL"),
		S3SignedURLTTL:   time.Duration(optInt("S3_SIGNED_URL_TTL", 900)) * time.Second,

		MLServiceURL:     req("ML_SERVICE_URL"),
		MLServiceTimeout: time.Duration(optInt("ML_SERVICE_TIMEOUT_MS", 30000)) * time.Millisecond,
		MLServiceAPIKey:  os.Getenv("ML_SERVICE_API_KEY"),

		MaxUploadSizeBytes: optInt64("MAX_UPLOAD_SIZE_BYTES", 15_728_640),
		AllowedImageMIME:   csv("ALLOWED_IMAGE_MIME", "image/jpeg,image/png,image/webp"),

		RateLimitWindow: time.Duration(optInt("RATE_LIMIT_WINDOW_MS", 60000)) * time.Millisecond,
		RateLimitMax:    optInt("RATE_LIMIT_MAX", 120),

		CORSOrigins: csv("CORS_ORIGINS", "http://localhost:3000"),
	}

	switch cfg.GoEnv {
	case "development", "test", "production":
	default:
		problems = append(problems, "GO_ENV must be one of development, test, production")
	}

	if len(cfg.JWTAccessSecret) < 8 && cfg.JWTAccessSecret != "" {
		problems = append(problems, "JWT_ACCESS_SECRET must be at least 8 characters")
	}
	if len(cfg.JWTRefreshSecret) < 8 && cfg.JWTRefreshSecret != "" {
		problems = append(problems, "JWT_REFRESH_SECRET must be at least 8 characters")
	}

	if len(problems) > 0 {
		return nil, fmt.Errorf("invalid environment configuration:\n  - %s", strings.Join(problems, "\n  - "))
	}

	return cfg, nil
}

// parseGoOrJSDuration accepts Go durations (15m, 30s) plus a "7d" days
// shorthand (used by JWT_REFRESH_TTL, matching the Node .env.example).
func parseGoOrJSDuration(v string) (time.Duration, error) {
	if strings.HasSuffix(v, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(v, "d"))
		if err != nil {
			return 0, err
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(v)
}
