package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/playperu/cityquiz/internal/handler"
)

type mockPinger struct {
	err error
}

func (m mockPinger) Ping(_ context.Context) error { return m.err }

func TestHealth(t *testing.T) {
	tests := []struct {
		name       string
		dbErr      error
		redisErr   error
		wantStatus int
		wantSQLite string
		wantRedis  string
	}{
		{
			name:       "both healthy",
			wantStatus: http.StatusOK,
			wantSQLite: "ok",
			wantRedis:  "ok",
		},
		{
			name:       "sqlite down",
			dbErr:      errors.New("database is locked"),
			wantStatus: http.StatusServiceUnavailable,
			wantSQLite: "error",
			wantRedis:  "ok",
		},
		{
			name:       "redis down",
			redisErr:   errors.New("connection refused"),
			wantStatus: http.StatusServiceUnavailable,
			wantSQLite: "ok",
			wantRedis:  "error",
		},
		{
			name:       "both down",
			dbErr:      errors.New("db down"),
			redisErr:   errors.New("redis down"),
			wantStatus: http.StatusServiceUnavailable,
			wantSQLite: "error",
			wantRedis:  "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handler.NewHealth(
				mockPinger{err: tt.dbErr},
				mockPinger{err: tt.redisErr},
				slog.Default(),
			)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			h.Routes().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var body map[string]struct {
				Status string `json:"status"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decoding response: %v", err)
			}

			if body["sqlite"].Status != tt.wantSQLite {
				t.Errorf("sqlite status = %q, want %q", body["sqlite"].Status, tt.wantSQLite)
			}
			if body["redis"].Status != tt.wantRedis {
				t.Errorf("redis status = %q, want %q", body["redis"].Status, tt.wantRedis)
			}
		})
	}
}
