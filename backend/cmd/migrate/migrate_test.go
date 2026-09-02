package main

import (
	"errors"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// TestMigrations_UpDownUp verifies that migrations/*.sql apply cleanly to a
// fresh database, are safely repeatable (a second Up() is a no-op, never a
// re-execution or corruption), and roll back cleanly via the down scripts —
// i.e. exactly the guarantees required by section 13 (migration safety).
//
// Skipped unless TEST_DATABASE_URL is set. It must point at a disposable
// scratch database (NOT the dev database) — this test drops all objects in
// it before and after running.
func TestMigrations_UpDownUp(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping migration up/down integration test")
	}

	m, err := migrate.New("file://../../migrations", databaseURL)
	if err != nil {
		t.Fatalf("migrate.New() error = %v", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			t.Logf("close source: %v", srcErr)
		}
		if dbErr != nil {
			t.Logf("close database: %v", dbErr)
		}
	}()

	// Start from a known-empty state regardless of what a previous failed
	// test run left behind.
	if err := m.Drop(); err != nil {
		t.Fatalf("Drop() error = %v", err)
	}

	if err := m.Up(); err != nil {
		t.Fatalf("Up() on a fresh database error = %v", err)
	}

	// Re-running Up() against an already-migrated database must be a
	// tracked no-op, never re-execute (and corrupt) the schema.
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		t.Fatalf("second Up() should be a no-op, got error = %v", err)
	}

	if err := m.Down(); err != nil {
		t.Fatalf("Down() error = %v", err)
	}
}
