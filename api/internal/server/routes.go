package server

import (
	"database/sql"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

func addRoutes(r chi.Router, logger *slog.Logger, db *sql.DB, rdb *redis.Client) {
	r.Get("/healthz", handleHealth(logger, db, rdb))
	r.Get("/ws/echo", handleWSEcho(logger))
}
