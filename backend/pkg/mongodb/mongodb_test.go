package mongodb

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestConnect_LiveDatabase verifies client creation + ping against a real
// MongoDB instance. Skipped unless TEST_MONGODB_URI is set.
func TestConnect_LiveDatabase(t *testing.T) {
	uri := os.Getenv("TEST_MONGODB_URI")
	if uri == "" {
		t.Skip("TEST_MONGODB_URI not set; skipping live MongoDB connectivity test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	client, db, err := Connect(ctx, uri, "legal_metrology_ml_test")
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Disconnect(ctx)

	if err := client.Ping(ctx, nil); err != nil {
		t.Fatalf("Ping() after Connect() error = %v", err)
	}
	if db.Name() == "" {
		t.Error("expected a non-empty database name")
	}
}

func TestConnect_InvalidURI_ReturnsError(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, _, err := Connect(ctx, "not-a-valid-mongo-uri", "fallback")
	if err == nil {
		t.Fatal("expected an error for a malformed connection string")
	}
}
