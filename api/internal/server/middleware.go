package server

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type ctxKey int

const (
	ctxKeyStore ctxKey = iota
	ctxKeyAdmin
)

func clientMiddleware(clients *Registry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slug := chi.URLParam(r, "client")
			if slug == "" {
				writeError(w, http.StatusNotFound, "client not found")
				return
			}

			store, err := clients.Get(r.Context(), slug)
			if err != nil {
				writeError(w, http.StatusNotFound, "client not found")
				return
			}

			ctx := context.WithValue(r.Context(), ctxKeyStore, Store(store))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func adminAuthMiddleware(admin AdminStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			ctx := context.WithValue(r.Context(), ctxKeyAdmin, sess)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func clientStore(r *http.Request) Store {
	return r.Context().Value(ctxKeyStore).(Store)
}

func adminFrom(r *http.Request) adminSession {
	return r.Context().Value(ctxKeyAdmin).(adminSession)
}
