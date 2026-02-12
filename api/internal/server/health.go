package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

func handleHealth(logger *slog.Logger, db *sql.DB, rdb *redis.Client) http.HandlerFunc {
	type result struct {
		Status string `json:"status"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		checks := map[string]result{
			"sqlite": {Status: "ok"},
			"redis":  {Status: "ok"},
		}
		status := http.StatusOK

		if err := db.PingContext(ctx); err != nil {
			logger.Error("health check failed", "name", "sqlite", "error", err)
			checks["sqlite"] = result{Status: "error"}
			status = http.StatusServiceUnavailable
		}

		if err := rdb.Ping(ctx).Err(); err != nil {
			logger.Error("health check failed", "name", "redis", "error", err)
			checks["redis"] = result{Status: "error"}
			status = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		json.NewEncoder(w).Encode(checks)
	}
}
