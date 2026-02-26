package server

import (
	"net/http"
)

func handleAdminLogout(admin AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(adminCookieName)
		if err == nil && cookie.Value != "" {
			admin.DeleteAdminSession(r.Context(), cookie.Value)
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
