package server

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/swaggest/swgui/v5emb"
)

func addRoutes(r chi.Router, logger *slog.Logger, admin AdminStore, clients *Registry, adminDB *sql.DB, spaDir string) {
	broker := NewBroker()

	r.Get("/openapi.json", handleOpenAPI())
	r.Mount("/docs", v5emb.New("CityQuest API", "/openapi.json", "/docs"))
	r.Get("/healthz", handleHealth(logger, adminDB))
	r.Get("/ws/echo", handleWSEcho(logger))

	// Player routes — {client} resolved by clientMiddleware.
	r.Route("/api/{client}", func(r chi.Router) {
		r.Use(clientMiddleware(clients))
		r.Get("/teams/{joinToken}", handleTeamLookup())
		r.Post("/join", handleJoin(broker))
		r.Get("/game/state", handleGameState())
		r.Post("/game/answer", handleAnswer(broker))
		r.Post("/game/unlock", handleUnlock(broker))
		r.Get("/game/events", handleEvents(broker))
	})

	// Admin auth — shared DB.
	r.Post("/api/admin/login", handleAdminLogin(admin))
	r.Post("/api/admin/logout", handleAdminLogout(admin))
	r.Get("/api/admin/me", handleAdminMe(admin))
	r.Get("/api/admin/clients", handleAdminListClients(admin))
	r.Post("/api/admin/clients", handleAdminCreateClient(admin, clients))

	// Admin scenarios — global, stored in admin DB.
	r.Route("/api/admin/scenarios", func(r chi.Router) {
		r.Use(adminAuthMiddleware(admin))
		r.Get("/", handleAdminListScenarios(admin))
		r.Post("/", handleAdminCreateScenario(admin))
		r.Get("/{id}", handleAdminGetScenario(admin))
		r.Put("/{id}", handleAdminUpdateScenario(admin))
		r.Delete("/{id}", handleAdminDeleteScenario(admin, clients))
	})

	// Admin games/teams — per-client, requires admin auth.
	r.Route("/api/admin/clients/{client}", func(r chi.Router) {
		r.Use(adminAuthMiddleware(admin))
		r.Use(clientMiddleware(clients))

		r.Get("/games", handleAdminListGames())
		r.Post("/games", handleAdminCreateGame(admin))
		r.Get("/games/{gameID}", handleAdminGetGame())
		r.Put("/games/{gameID}", handleAdminUpdateGame(admin))
		r.Delete("/games/{gameID}", handleAdminDeleteGame())
		r.Get("/games/{gameID}/status", handleAdminGameStatus())
		r.Get("/games/{gameID}/teams", handleAdminListTeams())
		r.Post("/games/{gameID}/teams", handleAdminCreateTeam())
		r.Put("/games/{gameID}/teams/{teamID}", handleAdminUpdateTeam())
		r.Delete("/games/{gameID}/teams/{teamID}", handleAdminDeleteTeam())
	})

	if spaDir != "" {
		if info, err := os.Stat(spaDir); err == nil && info.IsDir() {
			logger.Info("serving SPA", "dir", spaDir)
			r.NotFound(handleSPA(spaDir))
		}
	}
}
