package redisclient

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestConnect_LiveRedis verifies client creation + PING against a real Redis
// instance. Skipped unless TEST_REDIS_URL is set.
func TestConnect_LiveRedis(t *testing.T) {
	url := os.Getenv("TEST_REDIS_URL")
	if url == "" {
		t.Skip("TEST_REDIS_URL not set; skipping live Redis connectivity test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := Connect(ctx, url)
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer client.Close()

	if err := client.Ping(ctx).Err(); err != nil {
		t.Fatalf("Ping() after Connect() error = %v", err)
	}

	// Round-trip a key to confirm read/write, not just connectivity.
	key := "infra_test:redis_connect"
	if err := client.Set(ctx, key, "ok", time.Minute).Err(); err != nil {
		t.Fatalf("SET error = %v", err)
	}
	defer client.Del(ctx, key)

	val, err := client.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	if val != "ok" {
		t.Errorf("expected GET to return %q, got %q", "ok", val)
	}
}

func TestConnect_InvalidURL_ReturnsError(t *testing.T) {
	_, err := Connect(context.Background(), "not-a-valid-redis-url")
	if err == nil {
		t.Fatal("expected an error for a malformed connection string")
	}
}
