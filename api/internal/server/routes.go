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

		r.Route("/admin", func(r chi.Router) {
			r.Post("/login", handleAdminLogin(db))
			r.Post("/logout", handleAdminLogout(db))
			r.Get("/me", handleAdminMe(db))
			r.Get("/scenarios", handleAdminListScenarios(db))
			r.Post("/scenarios", handleAdminCreateScenario(db))
			r.Get("/scenarios/{id}", handleAdminGetScenario(db))
			r.Put("/scenarios/{id}", handleAdminUpdateScenario(db))
			r.Delete("/scenarios/{id}", handleAdminDeleteScenario(db))

			r.Get("/games", handleAdminListGames(db))
			r.Post("/games", handleAdminCreateGame(db))
			r.Get("/games/{gameID}", handleAdminGetGame(db))
			r.Put("/games/{gameID}", handleAdminUpdateGame(db))
			r.Delete("/games/{gameID}", handleAdminDeleteGame(db))
			r.Get("/games/{gameID}/teams", handleAdminListTeams(db))
			r.Post("/games/{gameID}/teams", handleAdminCreateTeam(db))
			r.Put("/games/{gameID}/teams/{teamID}", handleAdminUpdateTeam(db))
			r.Delete("/games/{gameID}/teams/{teamID}", handleAdminDeleteTeam(db))
		})
	})

	if spaDir != "" {
		if info, err := os.Stat(spaDir); err == nil && info.IsDir() {
			logger.Info("serving SPA", "dir", spaDir)
			r.NotFound(handleSPA(spaDir))
		}
	}
}
