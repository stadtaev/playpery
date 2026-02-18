package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
)

type playerSession struct {
	PlayerID string
	TeamID   string
	GameID   string
}

var errNoSession = errors.New("no valid session")

// playerFromToken looks up a player session by token (session_id).
func playerFromToken(db *sql.DB, token string) (playerSession, error) {
	var s playerSession
	err := db.QueryRow(`
		SELECT p.id, p.team_id, t.game_id
		FROM players p
		JOIN teams t ON t.id = p.team_id
		WHERE p.session_id = ?
	`, token).Scan(&s.PlayerID, &s.TeamID, &s.GameID)
	if errors.Is(err, sql.ErrNoRows) {
		return s, errNoSession
	}
	return s, err
}

// playerFromRequest extracts the Bearer token from the Authorization header
// and looks up the player session.
func playerFromRequest(r *http.Request, db *sql.DB) (playerSession, error) {
	auth := r.Header.Get("Authorization")
	token, found := strings.CutPrefix(auth, "Bearer ")
	if !found || token == "" {
		return playerSession{}, errNoSession
	}
	return playerFromToken(db, token)
}
