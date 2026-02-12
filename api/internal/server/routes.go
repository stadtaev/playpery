package server

import (
	"database/sql"
	"log/slog"

	"github.com/go-chi/chi/v5"
)

func addRoutes(r chi.Router, logger *slog.Logger, db *sql.DB) {
	r.Get("/openapi.json", handleOpenAPI())
	r.Get("/docs", handleSwaggerUI())
	r.Get("/healthz", handleHealth(logger, db))
	r.Get("/ws/echo", handleWSEcho(logger))
}
