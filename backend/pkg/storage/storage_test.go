package storage

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestObjectStorage_LiveMinIO exercises EnsureBucket/Ping/Put/Get/Delete
// against a real S3-compatible endpoint (MinIO in dev). Skipped unless
// TEST_S3_ENDPOINT is set.
func TestObjectStorage_LiveMinIO(t *testing.T) {
	endpoint := os.Getenv("TEST_S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("TEST_S3_ENDPOINT not set; skipping live object storage connectivity test")
	}

	bucket := os.Getenv("TEST_S3_BUCKET")
	if bucket == "" {
		bucket = "legal-metrology-test"
	}

	store, err := New(Config{
		Endpoint:       endpoint,
		Region:         envOrDefault("TEST_S3_REGION", "us-east-1"),
		AccessKey:      envOrDefault("TEST_S3_ACCESS_KEY", "minioadmin"),
		SecretKey:      envOrDefault("TEST_S3_SECRET_KEY", "minioadmin"),
		Bucket:         bucket,
		UseSSL:         os.Getenv("TEST_S3_USE_SSL") == "true",
		ForcePathStyle: true,
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := store.EnsureBucket(ctx); err != nil {
		t.Fatalf("EnsureBucket() error = %v", err)
	}
	if err := store.Ping(ctx); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	key := BuildObjectKey("infra-test", "connectivity.txt")
	body := []byte("infra test payload")

	if err := store.Put(ctx, key, body, "text/plain"); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	defer store.Delete(ctx, key)

	got, err := store.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if string(got) != string(body) {
		t.Errorf("expected round-tripped body %q, got %q", body, got)
	}

	if _, err := store.SignedDownloadURL(ctx, key, time.Minute); err != nil {
		t.Errorf("SignedDownloadURL() error = %v", err)
	}

	if err := store.Delete(ctx, key); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestSHA256Hex(t *testing.T) {
	got := SHA256Hex([]byte("hello"))
	want := "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if got != want {
		t.Errorf("SHA256Hex(%q) = %s, want %s", "hello", got, want)
	}
}

func TestBuildObjectKey_SanitizesFilename(t *testing.T) {
	key := BuildObjectKey("inspections/123", "my label image!.jpg")
	if key == "" {
		t.Fatal("expected a non-empty key")
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
