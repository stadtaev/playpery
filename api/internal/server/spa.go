package server

import (
	"net/http"
	"os"
	"path/filepath"
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
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fall back to index.html for SPA routes.
		http.ServeFile(w, r, filepath.Join(dir, "index.html"))
	}
}
