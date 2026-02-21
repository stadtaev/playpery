package server

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/swaggest/swgui/v5emb"
)

func addRoutes(r chi.Router, logger *slog.Logger, store Store, db *sql.DB, spaDir string) {
	broker := NewBroker()

	r.Get("/openapi.json", handleOpenAPI())
	r.Mount("/docs", v5emb.New("CityQuiz API", "/openapi.json", "/docs"))
	r.Get("/healthz", handleHealth(logger, db))
	r.Get("/ws/echo", handleWSEcho(logger))

	r.Route("/api", func(r chi.Router) {
		r.Get("/teams/{joinToken}", handleTeamLookup(store))
		r.Post("/join", handleJoin(store, broker))
		r.Get("/game/state", handleGameState(store))
		r.Post("/game/answer", handleAnswer(store, broker))
		r.Get("/game/events", handleEvents(store, broker))

		r.Route("/admin", func(r chi.Router) {
			r.Post("/login", handleAdminLogin(store))
			r.Post("/logout", handleAdminLogout(store))
			r.Get("/me", handleAdminMe(store))
			r.Get("/scenarios", handleAdminListScenarios(store))
			r.Post("/scenarios", handleAdminCreateScenario(store))
			r.Get("/scenarios/{id}", handleAdminGetScenario(store))
			r.Put("/scenarios/{id}", handleAdminUpdateScenario(store))
			r.Delete("/scenarios/{id}", handleAdminDeleteScenario(store))

			r.Get("/games", handleAdminListGames(store))
			r.Post("/games", handleAdminCreateGame(store))
			r.Get("/games/{gameID}", handleAdminGetGame(store))
			r.Put("/games/{gameID}", handleAdminUpdateGame(store))
			r.Delete("/games/{gameID}", handleAdminDeleteGame(store))
			r.Get("/games/{gameID}/teams", handleAdminListTeams(store))
			r.Post("/games/{gameID}/teams", handleAdminCreateTeam(store))
			r.Put("/games/{gameID}/teams/{teamID}", handleAdminUpdateTeam(store))
			r.Delete("/games/{gameID}/teams/{teamID}", handleAdminDeleteTeam(store))
		})
	})

	if spaDir != "" {
		if info, err := os.Stat(spaDir); err == nil && info.IsDir() {
			logger.Info("serving SPA", "dir", spaDir)
			r.NotFound(handleSPA(spaDir))
		}
	}
}
