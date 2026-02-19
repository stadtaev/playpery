package server

import (
	"database/sql"
	"errors"
	"net/http"
)

type adminSession struct {
	AdminID string
	Email   string
}

var errNoAdminSession = errors.New("no valid admin session")

const adminCookieName = "admin_session"

// adminFromRequest reads the admin_session cookie and looks up the admin session.
func adminFromRequest(r *http.Request, db *sql.DB) (adminSession, error) {
	cookie, err := r.Cookie(adminCookieName)
	if err != nil || cookie.Value == "" {
		return adminSession{}, errNoAdminSession
	}

	var s adminSession
	err = db.QueryRowContext(r.Context(), `
		SELECT a.id, a.email
		FROM admin_sessions s
		JOIN admins a ON a.id = s.admin_id
		WHERE s.id = ?
	`, cookie.Value).Scan(&s.AdminID, &s.Email)
	if errors.Is(err, sql.ErrNoRows) {
		return adminSession{}, errNoAdminSession
	}
	return s, err
}
