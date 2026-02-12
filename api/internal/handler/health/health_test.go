package health_test

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/playperu/cityquiz/internal/handler/health"
)

type mockChecker struct{ err error }

func (m mockChecker) Check(_ context.Context) error { return m.err }

func TestHandler(t *testing.T) {
	tests := []struct {
		name       string
		checks     map[string]health.Checker
		wantStatus int
		wantBody   map[string]string
	}{
		{
			name: "all healthy",
			checks: map[string]health.Checker{
				"sqlite": mockChecker{},
				"redis":  mockChecker{},
			},
			wantStatus: http.StatusOK,
			wantBody:   map[string]string{"sqlite": "ok", "redis": "ok"},
		},
		{
			name: "sqlite down",
			checks: map[string]health.Checker{
				"sqlite": mockChecker{err: errors.New("locked")},
				"redis":  mockChecker{},
			},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   map[string]string{"sqlite": "error", "redis": "ok"},
		},
		{
			name: "redis down",
			checks: map[string]health.Checker{
				"sqlite": mockChecker{},
				"redis":  mockChecker{err: errors.New("refused")},
			},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   map[string]string{"sqlite": "ok", "redis": "error"},
		},
		{
			name: "both down",
			checks: map[string]health.Checker{
				"sqlite": mockChecker{err: errors.New("db")},
				"redis":  mockChecker{err: errors.New("cache")},
			},
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   map[string]string{"sqlite": "error", "redis": "error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := health.NewHandler(slog.Default(), tt.checks)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			h.Routes().ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}

			var body map[string]struct{ Status string }
			if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
				t.Fatalf("decoding response: %v", err)
			}

			for name, want := range tt.wantBody {
				if got := body[name].Status; got != want {
					t.Errorf("%s status = %q, want %q", name, got, want)
				}
			}
		})
	}
}
