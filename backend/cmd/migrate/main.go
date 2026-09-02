// Migration runner. Wraps golang-migrate so `migrations/*.sql` are applied
// with a tracked schema_migrations table (safe to run repeatedly — already
// applied versions are skipped, never re-executed) instead of the old
// "psql -f migrations/0001_init.up.sql" flow, which had no version tracking
// and would fail (or silently re-run) if invoked twice.
//
// Uses golang-migrate's own postgres driver (lib/pq) purely as a migration
// tool dependency — independent of the pgx/v5 pool the API and worker use for
// runtime queries.
//
// Usage:
//
//	go run ./cmd/migrate up            # apply all pending migrations
//	go run ./cmd/migrate down          # roll back one migration
//	go run ./cmd/migrate down-all      # roll back all migrations (dev only)
//	go run ./cmd/migrate version       # print current schema version
//	go run ./cmd/migrate force VERSION # mark a dirty version as clean (recovery only)
package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: migrate <up|down|down-all|version|force VERSION>")
		os.Exit(1)
	}
	cmdName := os.Args[1]

	// Deliberately reads DATABASE_URL directly rather than internal/config:
	// a migration run needs only the database connection, not the full set
	// of app secrets (JWT, S3, ML service) config.Load() requires.
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	m, err := migrate.New("file://"+migrationsPath, databaseURL)
	if err != nil {
		fmt.Fprintln(os.Stderr, "migrate init:", err)
		os.Exit(1)
	}

	switch cmdName {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	case "down-all":
		err = m.Down()
	case "version":
		v, dirty, verr := m.Version()
		if verr != nil {
			err = verr
		} else {
			fmt.Printf("version=%d dirty=%v\n", v, dirty)
		}
	case "force":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "usage: migrate force VERSION")
			os.Exit(1)
		}
		var version int
		if _, serr := fmt.Sscanf(os.Args[2], "%d", &version); serr != nil {
			fmt.Fprintln(os.Stderr, "invalid version:", os.Args[2])
			os.Exit(1)
		}
		err = m.Force(version)
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", cmdName)
		os.Exit(1)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		fmt.Fprintln(os.Stderr, "migrate", cmdName, ":", err)
		os.Exit(1)
	}
	fmt.Println("migrate", cmdName, "complete")
}
