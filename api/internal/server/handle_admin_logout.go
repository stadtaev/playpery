package server

import (
	"database/sql"
	"net/http"
)

func handleAdminLogout(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(adminCookieName)
		if err == nil && cookie.Value != "" {
			db.ExecContext(r.Context(), `DELETE FROM admin_sessions WHERE id = ?`, cookie.Value)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     adminCookieName,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
