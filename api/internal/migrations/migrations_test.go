package migrations_test

import (
	"context"
	"testing"

	"github.com/playperu/cityquiz/internal/database"
	"github.com/playperu/cityquiz/internal/migrations"
)

func TestMigrations(t *testing.T) {
	db, err := database.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	if err := migrations.Run(db); err != nil {
		t.Fatalf("running migrations: %v", err)
	}

	// Verify all tables exist by querying sqlite_master.
	want := []string{"clients", "scenarios", "games", "teams", "players", "stage_results", "admins", "admin_sessions"}

	for _, table := range want {
		var name string
		err := db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestMigrationsIdempotent(t *testing.T) {
	db, err := database.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()

	if err := migrations.Run(db); err != nil {
		t.Fatalf("first run: %v", err)
	}
	if err := migrations.Run(db); err != nil {
		t.Fatalf("second run (should be no-op): %v", err)
	}
}
