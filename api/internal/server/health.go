package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// HealthCheckResult represents the status of a single health check.
type HealthCheckResult struct {
	Status string `json:"status" enum:"ok,error"`
}

// HealthResponse is the top-level response from GET /healthz.
type HealthResponse struct {
	SQLite HealthCheckResult `json:"sqlite"`
}

func handleHealth(logger *slog.Logger, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		resp := HealthResponse{
			SQLite: HealthCheckResult{Status: "ok"},
		}
		status := http.StatusOK

		if err := db.PingContext(ctx); err != nil {
			logger.Error("health check failed", "name", "sqlite", "error", err)
			resp.SQLite = HealthCheckResult{Status: "error"}
			status = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(resp)
	}
}
