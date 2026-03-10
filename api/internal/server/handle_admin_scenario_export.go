package server

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
)

const maxImportSize = 32 << 20 // 32 MB

func handleAdminExportScenario(admin AdminStore, dataDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		scenario, err := admin.GetScenario(r.Context(), id)
		if err != nil {
			if err.Error() == "scenario not found" || err == ErrNotFound {
				writeError(w, http.StatusNotFound, "scenario not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		// Build AdminScenarioRequest for the JSON block.
		req := AdminScenarioRequest{
			Name:        scenario.Name,
			City:        scenario.City,
			Description: scenario.Description,
			Mode:        scenario.Mode,
			Stages:      make([]AdminStage, len(scenario.Stages)),
		}
		copy(req.Stages, scenario.Stages)

		// Convert file paths to data URIs in the request copy.
		for i := range req.Stages {
			req.Stages[i].ClueImage = imageToDataURI(dataDir, req.Stages[i].ClueImage)
			req.Stages[i].QuestionImage = imageToDataURI(dataDir, req.Stages[i].QuestionImage)
			for j := range req.Stages[i].FunFacts {
				req.Stages[i].FunFacts[j].Image = imageToDataURI(dataDir, req.Stages[i].FunFacts[j].Image)
			}
		}

		jsonBytes, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		md := buildExportMarkdown(req, string(jsonBytes))
		filename := slugify(scenario.Name) + ".md"

		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, md)
	}
}

func handleAdminImportScenario(admin AdminStore, dataDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxImportSize+1024)

		file, _, err := r.FormFile("file")
		if err != nil {
			writeError(w, http.StatusBadRequest, "file is required")
			return
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read file")
			return
		}

		content := string(data)

		// Extract JSON from <!-- SCENARIO_JSON ... -->
		const startMarker = "<!-- SCENARIO_JSON"
		const endMarker = "-->"

		startIdx := strings.Index(content, startMarker)
		if startIdx == -1 {
			writeError(w, http.StatusBadRequest, "no SCENARIO_JSON block found")
			return
		}
		jsonStart := startIdx + len(startMarker)

		endIdx := strings.Index(content[jsonStart:], endMarker)
		if endIdx == -1 {
			writeError(w, http.StatusBadRequest, "malformed SCENARIO_JSON block")
			return
		}
		jsonStr := strings.TrimSpace(content[jsonStart : jsonStart+endIdx])

		var req AdminScenarioRequest
		if err := json.Unmarshal([]byte(jsonStr), &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON in SCENARIO_JSON block")
			return
		}

		// Convert data URIs to files.
		for i := range req.Stages {
			req.Stages[i].ClueImage = dataURIToFile(dataDir, req.Stages[i].ClueImage)
			req.Stages[i].QuestionImage = dataURIToFile(dataDir, req.Stages[i].QuestionImage)
			for j := range req.Stages[i].FunFacts {
				req.Stages[i].FunFacts[j].Image = dataURIToFile(dataDir, req.Stages[i].FunFacts[j].Image)
			}
		}

		if msg := req.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		// Check name collision.
		scenarios, err := admin.ListScenarios(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		for _, s := range scenarios {
			if strings.EqualFold(s.Name, req.Name) {
				writeError(w, http.StatusConflict, "scenario with this name already exists")
				return
			}
		}

		scenario, err := admin.CreateScenario(r.Context(), req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusCreated, scenario)
	}
}

// imageToDataURI reads an image file from disk and returns a data URI, or empty string.
func imageToDataURI(dataDir, path string) string {
	if path == "" || strings.HasPrefix(path, "data:") {
		return path
	}

	// path is like "/uploads/abc123.jpg"
	filename := strings.TrimPrefix(path, "/uploads/")
	if filename == path {
		return path // not an uploads path
	}

	fullPath := filepath.Join(dataDir, "uploads", filename)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "" // file not found, skip
	}

	mime := "image/jpeg"
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".png":
		mime = "image/png"
	case ".webp":
		mime = "image/webp"
	}

	return fmt.Sprintf("data:%s;base64,%s", mime, base64.StdEncoding.EncodeToString(data))
}

// dataURIToFile decodes a data URI, saves to disk, returns "/uploads/..." path.
func dataURIToFile(dataDir, uri string) string {
	if uri == "" || !strings.HasPrefix(uri, "data:") {
		return uri
	}

	// Parse "data:image/jpeg;base64,/9j/..."
	commaIdx := strings.Index(uri, ",")
	if commaIdx == -1 {
		return ""
	}

	header := uri[:commaIdx] // "data:image/jpeg;base64"
	encoded := uri[commaIdx+1:]

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return ""
	}

	ext := ".jpg"
	if strings.Contains(header, "image/png") {
		ext = ".png"
	} else if strings.Contains(header, "image/webp") {
		ext = ".webp"
	}

	uploadDir := filepath.Join(dataDir, "uploads")
	os.MkdirAll(uploadDir, 0o755)

	nameBytes := make([]byte, 16)
	rand.Read(nameBytes)
	name := hex.EncodeToString(nameBytes) + ext

	if err := os.WriteFile(filepath.Join(uploadDir, name), data, 0o644); err != nil {
		return ""
	}

	return "/uploads/" + name
}

var nonAlphaNum = regexp.MustCompile(`[^a-z0-9]+`)

// slugify converts a name to a URL-safe slug.
func slugify(name string) string {
	s := strings.ToLower(name)
	slug := nonAlphaNum.ReplaceAllString(s, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "scenario"
	}
	return slug
}

// buildExportMarkdown assembles the human-readable markdown + JSON block.
func buildExportMarkdown(req AdminScenarioRequest, jsonBlock string) string {
	var b strings.Builder

	b.WriteString("# ")
	b.WriteString(req.Name)
	b.WriteString("\n\n")

	b.WriteString("- **City:** ")
	b.WriteString(req.City)
	b.WriteString("\n")

	b.WriteString("- **Mode:** ")
	b.WriteString(req.Mode)
	b.WriteString("\n")

	if req.Description != "" {
		b.WriteString("- **Description:** ")
		b.WriteString(req.Description)
		b.WriteString("\n")
	}

	for _, stage := range req.Stages {
		b.WriteString("\n---\n\n")
		b.WriteString(fmt.Sprintf("## Stage %d — %s\n\n", stage.StageNumber, stage.Location))

		if stage.Clue != "" {
			b.WriteString("**Clue:** ")
			b.WriteString(stage.Clue)
			b.WriteString("\n\n")
		}
		if stage.ClueImage != "" {
			b.WriteString(fmt.Sprintf("![clue](%s)\n\n", stage.ClueImage))
		}

		if stage.Question != "" {
			b.WriteString("**Question:** ")
			b.WriteString(stage.Question)
			b.WriteString("\n\n")
		}
		if stage.QuestionImage != "" {
			b.WriteString(fmt.Sprintf("![question](%s)\n\n", stage.QuestionImage))
		}

		if stage.CorrectAnswer != "" {
			b.WriteString("**Answer:** ")
			b.WriteString(stage.CorrectAnswer)
			b.WriteString("\n\n")
		}

		if stage.Lat != 0 || stage.Lng != 0 {
			b.WriteString(fmt.Sprintf("**Coordinates:** %.6f, %.6f\n\n", stage.Lat, stage.Lng))
		}

		if len(stage.FunFacts) > 0 {
			b.WriteString("### Fun Facts\n\n")
			for i, ff := range stage.FunFacts {
				b.WriteString(fmt.Sprintf("%d. %s\n", i+1, ff.Text))
				if ff.Image != "" {
					b.WriteString(fmt.Sprintf("\n   ![](%s)\n", ff.Image))
				}
				b.WriteString("\n")
			}
		}
	}

	b.WriteString("---\n\n")
	b.WriteString("<!-- SCENARIO_JSON\n")
	b.WriteString(jsonBlock)
	b.WriteString("\n-->\n")

	return b.String()
}
