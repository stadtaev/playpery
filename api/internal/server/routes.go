package server

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/swaggest/swgui/v5emb"
)

func addRoutes(r chi.Router, logger *slog.Logger, db *sql.DB, spaDir string) {
	broker := NewBroker()

	r.Get("/openapi.json", handleOpenAPI())
	r.Mount("/docs", v5emb.New("CityQuiz API", "/openapi.json", "/docs"))
	r.Get("/healthz", handleHealth(logger, db))
	r.Get("/ws/echo", handleWSEcho(logger))

	r.Route("/api", func(r chi.Router) {
		r.Get("/teams/{joinToken}", handleTeamLookup(db))
		r.Post("/join", handleJoin(db, broker))
		r.Get("/game/state", handleGameState(db))
		r.Post("/game/answer", handleAnswer(db, broker))
		r.Get("/game/events", handleEvents(db, broker))
	})

	if spaDir != "" {
		if info, err := os.Stat(spaDir); err == nil && info.IsDir() {
			logger.Info("serving SPA", "dir", spaDir)
			r.NotFound(handleSPA(spaDir))
		}
	}
}
