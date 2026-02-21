package server

import (
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

func playerFromRequest(r *http.Request, store Store) (playerSession, error) {
	auth := r.Header.Get("Authorization")
	token, found := strings.CutPrefix(auth, "Bearer ")
	if !found || token == "" {
		return playerSession{}, errNoSession
	}
	return store.PlayerFromToken(r.Context(), token)
}
