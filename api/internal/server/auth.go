package server

import (
	"errors"
	"net/http"
	"strings"
)

var errNoSession = errors.New("no valid session")

func playerFromRequest(r *http.Request) (sessionInfo, error) {
	auth := r.Header.Get("Authorization")
	token, found := strings.CutPrefix(auth, "Bearer ")
	if !found || token == "" {
		return sessionInfo{}, errNoSession
	}
	return clientStore(r).PlayerFromToken(r.Context(), token)
}
