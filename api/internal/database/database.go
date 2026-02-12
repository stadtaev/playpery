package database

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/tursodatabase/go-libsql"
)

// Open creates a SQLite connection via libSQL and configures it for
// concurrent use: WAL journal mode, 5 s busy timeout, foreign keys enabled.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("libsql", "file:"+path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// libSQL rejects Exec for PRAGMAs that return rows, but some PRAGMAs
	// (like foreign_keys=ON) return nothing. Use QueryContext and drain rows
	// to handle both cases uniformly.
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	}
	for _, p := range pragmas {
		rows, err := db.QueryContext(ctx, p)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("executing %s: %w", p, err)
		}
		rows.Close()
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return db, nil
}
