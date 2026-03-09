package server

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const maxUploadSize = 10 << 20 // 10 MB

var allowedMIME = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

func handleUpload(dataDir string) http.HandlerFunc {
	uploadDir := filepath.Join(dataDir, "uploads")

	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize+1024) // small margin for multipart headers

		file, header, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "file is required")
			return
		}
		defer file.Close()

		if header.Size > maxUploadSize {
			writeError(w, http.StatusRequestEntityTooLarge, "file too large (max 10 MB)")
			return
		}

		ct := header.Header.Get("Content-Type")
		// Some browsers send "image/jpeg; charset=..." — strip params.
		ct = strings.SplitN(ct, ";", 2)[0]
		ct = strings.TrimSpace(ct)

		ext, ok := allowedMIME[ct]
		if !ok {
			writeError(w, http.StatusBadRequest, "only JPEG, PNG, and WebP images are allowed")
			return
		}

		if err := os.MkdirAll(uploadDir, 0o755); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		nameBytes := make([]byte, 16)
		rand.Read(nameBytes)
		name := hex.EncodeToString(nameBytes) + ext

		dst, err := os.Create(filepath.Join(uploadDir, name))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"url": fmt.Sprintf("/uploads/%s", name),
		})
	}
}
