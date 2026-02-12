package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Pinger interface {
	Ping(ctx context.Context) error
}

type Health struct {
	db     Pinger
	redis  Pinger
	logger *slog.Logger
}

func NewHealth(db, redis Pinger, logger *slog.Logger) *Health {
	return &Health{db: db, redis: redis, logger: logger}
}

func (h *Health) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.check)
	return r
}

type checkResult struct {
	Status string `json:"status"`
}

func (h *Health) check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result := map[string]checkResult{
		"sqlite": {Status: "ok"},
		"redis":  {Status: "ok"},
	}
	status := http.StatusOK

	if err := h.db.Ping(ctx); err != nil {
		h.logger.Error("sqlite health check failed", "error", err)
		result["sqlite"] = checkResult{Status: "error"}
		status = http.StatusServiceUnavailable
	}

	if err := h.redis.Ping(ctx); err != nil {
		h.logger.Error("redis health check failed", "error", err)
		result["redis"] = checkResult{Status: "error"}
		status = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(result)
}
