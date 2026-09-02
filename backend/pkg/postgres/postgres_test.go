package postgres

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestConnect_LiveDatabase verifies pool creation + ping against a real
// PostgreSQL instance. Skipped unless TEST_DATABASE_URL is set (e.g. pointed
// at the docker-compose postgres service) so `go test ./...` stays infra-free
// by default.
func TestConnect_LiveDatabase(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping live PostgreSQL connectivity test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := Connect(ctx, url)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		t.Fatalf("Ping() after Connect() error = %v", err)
	}
}

func TestConnect_InvalidURL_ReturnsError(t *testing.T) {
	_, err := Connect(context.Background(), "not-a-valid-postgres-url")
	if err == nil {
		t.Fatal("expected an error for a malformed connection string")
	}
}
