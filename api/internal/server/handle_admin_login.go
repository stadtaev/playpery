package server

import (
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

func handleAdminLogin(admin AdminAuth) http.HandlerFunc {
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

		adminID, passwordHash, err := admin.AdminByEmail(r.Context(), req.Email)
		if errors.Is(err, ErrNotFound) {
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

		sessionID, err := admin.CreateAdminSession(r.Context(), adminID)
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

func handleAdminMe(admin AdminAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(adminCookieName)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		sess, err := admin.AdminFromSession(r.Context(), cookie.Value)
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

func handleAdminListClients(admin AdminAuth) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(adminCookieName)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		if _, err := admin.AdminFromSession(r.Context(), cookie.Value); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		clients, err := admin.ListClients(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if clients == nil {
			clients = []ClientInfo{}
		}

		writeJSON(w, http.StatusOK, clients)
	}
}

type CreateClientRequest struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

func handleAdminCreateClient(admin AdminAuth, clients *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(adminCookieName)
		if err != nil || cookie.Value == "" {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		if _, err := admin.AdminFromSession(r.Context(), cookie.Value); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		var req CreateClientRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		req.Slug = strings.TrimSpace(req.Slug)
		req.Name = strings.TrimSpace(req.Name)
		if req.Slug == "" || req.Name == "" {
			writeError(w, http.StatusBadRequest, "slug and name are required")
			return
		}

		if err := admin.CreateClient(r.Context(), req.Slug, req.Name); err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				writeError(w, http.StatusConflict, "client slug already exists")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		if _, err := clients.Create(r.Context(), req.Slug); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusCreated, ClientInfo{Slug: req.Slug, Name: req.Name})
	}
}
