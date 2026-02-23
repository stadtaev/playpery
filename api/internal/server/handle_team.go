package server

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type TeamLookupResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	GameName string `json:"gameName"`
	Role     string `json:"role"`
	GameID   string `json:"-"`
}

func handleTeamLookup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := chi.URLParam(r, "joinToken")
		store := clientStore(r)

		resp, err := store.TeamLookup(r.Context(), token)
		if errors.Is(err, ErrNotFound) {
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
