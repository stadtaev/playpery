package server

import (
	"database/sql"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/swaggest/swgui/v5emb"
)

func addRoutes(r chi.Router, logger *slog.Logger, db *sql.DB) {
	r.Get("/openapi.json", handleOpenAPI())
	r.Mount("/docs", v5emb.New("CityQuiz API", "/openapi.json", "/docs"))
	r.Get("/healthz", handleHealth(logger, db))
	r.Get("/ws/echo", handleWSEcho(logger))
}
