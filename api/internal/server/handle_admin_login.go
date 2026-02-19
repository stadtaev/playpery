package server

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// AdminLoginRequest is the request body for POST /api/admin/login.
type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AdminMeResponse is the response for GET /api/admin/me.
type AdminMeResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func handleAdminLogin(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AdminLoginRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.Email = strings.TrimSpace(strings.ToLower(req.Email))
		if req.Email == "" || req.Password == "" {
			writeError(w, http.StatusBadRequest, "email and password are required")
			return
		}

		var adminID, passwordHash string
		err := db.QueryRowContext(r.Context(), `
			SELECT id, password_hash FROM admins WHERE email = ?
		`, req.Email).Scan(&adminID, &passwordHash)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		// Create session.
		var sessionID string
		err = db.QueryRowContext(r.Context(), `
			INSERT INTO admin_sessions (admin_id)
			VALUES (?)
			RETURNING id
		`, adminID).Scan(&sessionID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     adminCookieName,
			Value:    sessionID,
			Path:     "/",
			MaxAge:   int(7 * 24 * time.Hour / time.Second),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		writeJSON(w, http.StatusOK, AdminMeResponse{
			ID:    adminID,
			Email: req.Email,
		})
	}
}

func handleAdminMe(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := adminFromRequest(r, db)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}
		writeJSON(w, http.StatusOK, AdminMeResponse{
			ID:    sess.AdminID,
			Email: sess.Email,
		})
	}
}
