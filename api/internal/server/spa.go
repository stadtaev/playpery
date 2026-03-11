package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// handleLanding serves a static landing page with appropriate cache headers.
func handleLanding(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, path)
	}
}

// handleSPA serves static files from dir, falling back to index.html
// for any path that doesn't match a real file (SPA client-side routing).
func handleSPA(dir string) http.HandlerFunc {
	fs := http.Dir(dir)
	fileServer := http.FileServer(fs)

	return func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the exact file.
		path := filepath.Join(dir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			setCacheHeaders(w, r.URL.Path)
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fall back to index.html for SPA routes — must not be cached.
		w.Header().Set("Cache-Control", "no-cache")
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	}
}

// setCacheHeaders sets Cache-Control based on asset type.
// Vite hashed assets (/assets/*) are immutable. Everything else gets short cache.
func setCacheHeaders(w http.ResponseWriter, urlPath string) {
	switch {
	case strings.HasPrefix(urlPath, "/assets/"):
		// Vite content-hashed files — immutable, 2 days.
		w.Header().Set("Cache-Control", "public, max-age=172800, immutable")
	case urlPath == "/sw.js":
		// Service worker — browser checks for updates on every navigation.
		w.Header().Set("Cache-Control", "no-cache")
	default:
		// Icons, manifest, fonts — short cache, revalidate.
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}
}
