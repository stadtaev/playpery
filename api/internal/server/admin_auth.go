package server

import (
	"errors"
	"net/http"
)

type adminSession struct {
	AdminID string
	Email   string
}

var errNoAdminSession = errors.New("no valid admin session")

const adminCookieName = "admin_session"

func adminFromRequest(r *http.Request, store Store) (adminSession, error) {
	cookie, err := r.Cookie(adminCookieName)
	if err != nil || cookie.Value == "" {
		return adminSession{}, errNoAdminSession
	}
	return store.AdminFromSession(r.Context(), cookie.Value)
}
