package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// AdminScenarioSummary is returned in the list endpoint.
type AdminScenarioSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	City        string `json:"city"`
	Description string `json:"description"`
	StageCount  int    `json:"stageCount"`
	CreatedAt   string `json:"createdAt"`
}

// AdminScenarioDetail is returned for a single scenario with stages.
type AdminScenarioDetail struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	City        string       `json:"city"`
	Description string       `json:"description"`
	Stages      []AdminStage `json:"stages"`
	CreatedAt   string       `json:"createdAt"`
}

// AdminStage represents a stage in a scenario.
type AdminStage struct {
	StageNumber   int     `json:"stageNumber"`
	Location      string  `json:"location"`
	Clue          string  `json:"clue"`
	Question      string  `json:"question"`
	CorrectAnswer string  `json:"correctAnswer"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
}

// AdminScenarioRequest is the request body for creating/updating a scenario.
type AdminScenarioRequest struct {
	Name        string       `json:"name"`
	City        string       `json:"city"`
	Description string       `json:"description"`
	Stages      []AdminStage `json:"stages"`
}

func (req *AdminScenarioRequest) validate() string {
	req.Name = strings.TrimSpace(req.Name)
	req.City = strings.TrimSpace(req.City)
	req.Description = strings.TrimSpace(req.Description)
	if req.Name == "" {
		return "name is required"
	}
	if req.City == "" {
		return "city is required"
	}
	if len(req.Stages) == 0 {
		return "at least one stage is required"
	}
	for i := range req.Stages {
		req.Stages[i].StageNumber = i + 1
		if strings.TrimSpace(req.Stages[i].Location) == "" {
			return "each stage must have a location"
		}
		if strings.TrimSpace(req.Stages[i].Question) == "" {
			return "each stage must have a question"
		}
		if strings.TrimSpace(req.Stages[i].CorrectAnswer) == "" {
			return "each stage must have a correctAnswer"
		}
	}
	return ""
}

func handleAdminListScenarios(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		rows, err := db.QueryContext(r.Context(), `
			SELECT id, name, city, COALESCE(description, ''), stages, created_at
			FROM scenarios
			ORDER BY created_at DESC
		`)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		defer rows.Close()

		var scenarios []AdminScenarioSummary
		for rows.Next() {
			var s AdminScenarioSummary
			var stagesJSON string
			if err := rows.Scan(&s.ID, &s.Name, &s.City, &s.Description, &stagesJSON, &s.CreatedAt); err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			var stages []json.RawMessage
			json.Unmarshal([]byte(stagesJSON), &stages)
			s.StageCount = len(stages)
			scenarios = append(scenarios, s)
		}

		if scenarios == nil {
			scenarios = []AdminScenarioSummary{}
		}
		writeJSON(w, http.StatusOK, scenarios)
	}
}

func handleAdminCreateScenario(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		var req AdminScenarioRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if msg := req.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		stagesJSON, _ := json.Marshal(req.Stages)

		var id, createdAt string
		err := db.QueryRowContext(r.Context(), `
			INSERT INTO scenarios (id, name, city, description, stages)
			VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?)
			RETURNING id, created_at
		`, req.Name, req.City, req.Description, string(stagesJSON)).Scan(&id, &createdAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusCreated, AdminScenarioDetail{
			ID:          id,
			Name:        req.Name,
			City:        req.City,
			Description: req.Description,
			Stages:      req.Stages,
			CreatedAt:   createdAt,
		})
	}
}

func handleAdminGetScenario(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		id := chi.URLParam(r, "id")

		var s AdminScenarioDetail
		var stagesJSON string
		err := db.QueryRowContext(r.Context(), `
			SELECT id, name, city, COALESCE(description, ''), stages, created_at
			FROM scenarios WHERE id = ?
		`, id).Scan(&s.ID, &s.Name, &s.City, &s.Description, &stagesJSON, &s.CreatedAt)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "scenario not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		json.Unmarshal([]byte(stagesJSON), &s.Stages)
		if s.Stages == nil {
			s.Stages = []AdminStage{}
		}
		writeJSON(w, http.StatusOK, s)
	}
}

func handleAdminUpdateScenario(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		id := chi.URLParam(r, "id")

		var req AdminScenarioRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if msg := req.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		stagesJSON, _ := json.Marshal(req.Stages)

		var createdAt string
		err := db.QueryRowContext(r.Context(), `
			UPDATE scenarios SET name = ?, city = ?, description = ?, stages = ?
			WHERE id = ?
			RETURNING created_at
		`, req.Name, req.City, req.Description, string(stagesJSON), id).Scan(&createdAt)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "scenario not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, AdminScenarioDetail{
			ID:          id,
			Name:        req.Name,
			City:        req.City,
			Description: req.Description,
			Stages:      req.Stages,
			CreatedAt:   createdAt,
		})
	}
}

func handleAdminDeleteScenario(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		id := chi.URLParam(r, "id")

		// Block deletion if games reference this scenario.
		var gameCount int
		err := db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM games WHERE scenario_id = ?
		`, id).Scan(&gameCount)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if gameCount > 0 {
			writeError(w, http.StatusConflict, "cannot delete scenario with existing games")
			return
		}

		result, err := db.ExecContext(r.Context(), `DELETE FROM scenarios WHERE id = ?`, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			writeError(w, http.StatusNotFound, "scenario not found")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
