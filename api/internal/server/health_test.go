package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"time"

	"github.com/playperu/cityquiz/internal/database"
	"github.com/redis/go-redis/v9"
)

func deadRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         "localhost:1",
		DialTimeout:  10 * time.Millisecond,
		ReadTimeout:  10 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
		MaxRetries:   0,
	})
}

func TestHandleHealth(t *testing.T) {
	// Real SQLite in-memory DB — lightweight, no mocks needed.
	db, err := database.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("opening db: %v", err)
	}
	defer db.Close()

	tests := []struct {
		name       string
		db         *sql.DB
		rdb        *redis.Client
		wantStatus int
		wantSQLite string
		wantRedis  string
	}{
		{
			name:       "sqlite ok redis down",
			db:         db,
			rdb:        deadRedis(),
			wantStatus: http.StatusServiceUnavailable,
			wantSQLite: "ok",
			wantRedis:  "error",
		},
		{
			name:       "sqlite ok (redis still down)",
			db:         db,
			rdb:        deadRedis(),
			wantStatus: http.StatusServiceUnavailable,
			wantSQLite: "ok",
			wantRedis:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handleHealth(slog.Default(), tt.db, tt.rdb)

			req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			rec := httptest.NewRecorder()
			h(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var body map[string]struct{ Status string }
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decoding: %v", err)
			}
			if got := body["sqlite"].Status; got != tt.wantSQLite {
				t.Errorf("sqlite = %q, want %q", got, tt.wantSQLite)
			}
			if got := body["redis"].Status; got != tt.wantRedis {
				t.Errorf("redis = %q, want %q", got, tt.wantRedis)
			}
		})
	}
}
