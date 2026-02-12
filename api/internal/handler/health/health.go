package health

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// Checker verifies that an infrastructure dependency is reachable.
type Checker interface {
	Check(ctx context.Context) error
}

type Handler struct {
	checks map[string]Checker
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger, checks map[string]Checker) *Handler {
	return &Handler{checks: checks, logger: logger}
}

func (h *Handler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/", h.check)
	return r
}

type result struct {
	Status string `json:"status"`
}

func (h *Handler) check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	results := make(map[string]result, len(h.checks))
	status := http.StatusOK

	for name, c := range h.checks {
		if err := c.Check(ctx); err != nil {
			h.logger.Error("health check failed", "name", name, "error", err)
			results[name] = result{Status: "error"}
			status = http.StatusServiceUnavailable
			continue
		}
		results[name] = result{Status: "ok"}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(results)
}
