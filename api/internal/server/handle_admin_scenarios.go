package server

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type AdminScenarioSummary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	City        string `json:"city"`
	Description string `json:"description"`
	StageCount  int    `json:"stageCount"`
	CreatedAt   string `json:"createdAt"`
}

type AdminScenarioDetail struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	City        string       `json:"city"`
	Description string       `json:"description"`
	Stages      []AdminStage `json:"stages"`
	CreatedAt   string       `json:"createdAt"`
}

type AdminStage struct {
	StageNumber   int     `json:"stageNumber"`
	Location      string  `json:"location"`
	Clue          string  `json:"clue"`
	Question      string  `json:"question"`
	CorrectAnswer string  `json:"correctAnswer"`
	Lat           float64 `json:"lat"`
	Lng           float64 `json:"lng"`
}

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

func handleAdminListScenarios(admin AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scenarios, err := admin.ListScenarios(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		if scenarios == nil {
			scenarios = []AdminScenarioSummary{}
		}
		writeJSON(w, http.StatusOK, scenarios)
	}
}

func handleAdminCreateScenario(admin AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AdminScenarioRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if msg := req.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		scenario, err := admin.CreateScenario(r.Context(), req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusCreated, scenario)
	}
}

func handleAdminGetScenario(admin AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		scenario, err := admin.GetScenario(r.Context(), id)
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "scenario not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, scenario)
	}
}

func handleAdminUpdateScenario(admin AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		scenario, err := admin.UpdateScenario(r.Context(), id, req)
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "scenario not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, scenario)
	}
}

func handleAdminDeleteScenario(admin AdminStore, clients *Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		hasGames, err := admin.ScenarioHasGames(r.Context(), id, clients)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if hasGames {
			writeError(w, http.StatusConflict, "cannot delete scenario with existing games")
			return
		}

		if err := admin.DeleteScenario(r.Context(), id); err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusNotFound, "scenario not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
