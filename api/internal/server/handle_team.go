package server

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type TeamLookupResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	GameName string `json:"gameName"`
}

func handleTeamLookup(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "joinToken")

		var resp TeamLookupResponse
		err := db.QueryRowContext(r.Context(), `
			SELECT t.id, t.name, s.name
			FROM teams t
			JOIN games g ON g.id = t.game_id
			JOIN scenarios s ON s.id = g.scenario_id
			WHERE t.join_token = ? AND g.status = 'active'
		`, token).Scan(&resp.ID, &resp.Name, &resp.GameName)

		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "team not found or game not active")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, resp)
	}
}
