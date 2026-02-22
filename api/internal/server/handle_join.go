package server

import (
	"errors"
	"net/http"
	"strings"
)

type JoinRequest struct {
	JoinToken  string `json:"joinToken"`
	PlayerName string `json:"playerName"`
}

type JoinResponse struct {
	Token    string `json:"token"`
	PlayerID string `json:"playerId"`
	TeamID   string `json:"teamId"`
	TeamName string `json:"teamName"`
}

func handleJoin(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req JoinRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.PlayerName = strings.TrimSpace(req.PlayerName)
		if req.PlayerName == "" || req.JoinToken == "" {
			writeError(w, http.StatusBadRequest, "joinToken and playerName are required")
			return
		}

		store := clientStore(r)

		team, err := store.TeamLookup(r.Context(), req.JoinToken)
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "team not found or game not active")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		playerID, sessionID, err := store.JoinTeam(r.Context(), team.GameID, team.ID, req.PlayerName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		broker.Publish(team.ID, SSEEvent{
			Type:       "player_joined",
			PlayerName: req.PlayerName,
		})

		writeJSON(w, http.StatusOK, JoinResponse{
			Token:    sessionID,
			PlayerID: playerID,
			TeamID:   team.ID,
			TeamName: team.Name,
		})
	}
}
