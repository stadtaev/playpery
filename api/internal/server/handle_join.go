package server

import (
	"database/sql"
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

func handleJoin(db *sql.DB, broker *Broker) http.HandlerFunc {
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

		// Look up team + verify game is active.
		var teamID, teamName string
		err := db.QueryRowContext(r.Context(), `
			SELECT t.id, t.name
			FROM teams t
			JOIN games g ON g.id = t.game_id
			WHERE t.join_token = ? AND g.status = 'active'
		`, req.JoinToken).Scan(&teamID, &teamName)

		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "team not found or game not active")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		// Insert player with random session_id.
		var playerID, sessionID string
		err = db.QueryRowContext(r.Context(), `
			INSERT INTO players (team_id, name, session_id)
			VALUES (?, ?, lower(hex(randomblob(16))))
			RETURNING id, session_id
		`, teamID, req.PlayerName).Scan(&playerID, &sessionID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		broker.Publish(teamID, SSEEvent{
			Type:       "player_joined",
			PlayerName: req.PlayerName,
		})

		writeJSON(w, http.StatusOK, JoinResponse{
			Token:    sessionID,
			PlayerID: playerID,
			TeamID:   teamID,
			TeamName: teamName,
		})
	}
}
