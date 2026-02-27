package server

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

type AdminGameSummary struct {
	ID                string `json:"id"`
	ScenarioID        string `json:"scenarioId"`
	ScenarioName      string `json:"scenarioName"`
	Status            string `json:"status"`
	Mode              string `json:"mode"`
	Supervised        bool   `json:"supervised"`
	TimerEnabled      bool   `json:"timerEnabled"`
	TimerMinutes      int    `json:"timerMinutes"`
	StageTimerMinutes int    `json:"stageTimerMinutes"`
	TeamCount         int    `json:"teamCount"`
	CreatedAt         string `json:"createdAt"`
}

type AdminGameDetail struct {
	ID                string          `json:"id"`
	ScenarioID        string          `json:"scenarioId"`
	ScenarioName      string          `json:"scenarioName"`
	Status            string          `json:"status"`
	Mode              string          `json:"mode"`
	Supervised        bool            `json:"supervised"`
	TimerEnabled      bool            `json:"timerEnabled"`
	TimerMinutes      int             `json:"timerMinutes"`
	StageTimerMinutes int             `json:"stageTimerMinutes"`
	StartedAt         *string         `json:"startedAt"`
	Teams             []AdminTeamItem `json:"teams"`
	CreatedAt         string          `json:"createdAt"`
}

type AdminTeamItem struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	JoinToken       string `json:"joinToken"`
	SupervisorToken string `json:"supervisorToken,omitempty"`
	GuideName       string `json:"guideName"`
	TeamSecret      int    `json:"teamSecret,omitempty"`
	PlayerCount     int    `json:"playerCount"`
	CreatedAt       string `json:"createdAt"`
}

type AdminGameRequest struct {
	ScenarioID        string `json:"scenarioId"`
	ScenarioName      string `json:"-"`  // set by handler after validation
	Mode              string `json:"-"`  // set by handler from scenario
	HasQuestions      bool   `json:"-"`  // set by handler from scenario
	Status            string `json:"status"`
	Supervised        bool   `json:"supervised"`
	TimerEnabled      bool   `json:"timerEnabled"`
	TimerMinutes      int    `json:"timerMinutes"`
	StageTimerMinutes int    `json:"stageTimerMinutes"`
}

type AdminTeamRequest struct {
	Name      string `json:"name"`
	JoinToken string `json:"joinToken"`
	GuideName string `json:"guideName"`
}

type AdminGameStatus struct {
	ID                string            `json:"id"`
	ScenarioName      string            `json:"scenarioName"`
	Status            string            `json:"status"`
	Mode              string            `json:"mode"`
	Supervised        bool              `json:"supervised"`
	TimerEnabled      bool              `json:"timerEnabled"`
	TimerMinutes      int               `json:"timerMinutes"`
	StageTimerMinutes int               `json:"stageTimerMinutes"`
	StartedAt         *string           `json:"startedAt"`
	TotalStages       int               `json:"totalStages"`
	Teams             []AdminTeamStatus `json:"teams"`
}

type AdminTeamStatus struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	GuideName       string              `json:"guideName"`
	CompletedStages int                 `json:"completedStages"`
	Players         []AdminPlayerStatus `json:"players"`
}

type AdminPlayerStatus struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	JoinedAt string `json:"joinedAt"`
}

var validGameStatuses = map[string]bool{
	"draft":  true,
	"active": true,
	"paused": true,
	"ended":  true,
}

func (req *AdminGameRequest) validate() string {
	req.ScenarioID = strings.TrimSpace(req.ScenarioID)
	req.Status = strings.TrimSpace(req.Status)
	if req.ScenarioID == "" {
		return "scenarioId is required"
	}
	if req.Status == "" {
		req.Status = "draft"
	}
	if !validGameStatuses[req.Status] {
		return "status must be draft, active, paused, or ended"
	}
	if req.TimerEnabled {
		if req.TimerMinutes <= 0 {
			req.TimerMinutes = 120
		}
		if req.StageTimerMinutes <= 0 {
			req.StageTimerMinutes = 10
		}
	} else {
		req.TimerMinutes = 0
		req.StageTimerMinutes = 0
	}
	return ""
}

func (req *AdminTeamRequest) validate() string {
	req.Name = strings.TrimSpace(req.Name)
	req.JoinToken = strings.TrimSpace(req.JoinToken)
	req.GuideName = strings.TrimSpace(req.GuideName)
	if req.Name == "" {
		return "name is required"
	}
	return ""
}

func generateJoinToken() string {
	b := make([]byte, 4)
	rand.Read(b)
	return "team-" + hex.EncodeToString(b)
}

func generateSupervisorToken() string {
	b := make([]byte, 4)
	rand.Read(b)
	return "super-" + hex.EncodeToString(b)
}

func handleAdminListGames() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := clientStore(r)

		games, err := store.ListGames(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		if games == nil {
			games = []AdminGameSummary{}
		}
		writeJSON(w, http.StatusOK, games)
	}
}

func handleAdminCreateGame(admin AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := clientStore(r)

		var req AdminGameRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if msg := req.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		scenario, err := admin.GetScenario(r.Context(), req.ScenarioID)
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusBadRequest, "scenario not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		req.ScenarioName = scenario.Name
		req.Mode = scenario.Mode
		req.HasQuestions = scenario.HasQuestions
		if req.Mode == "guided" {
			req.Supervised = true
		}

		game, err := store.CreateGame(r.Context(), req, scenario.Stages)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusCreated, game)
	}
}

func handleAdminGetGame() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := clientStore(r)
		gameID := chi.URLParam(r, "gameID")

		game, err := store.GetGame(r.Context(), gameID)
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, game)
	}
}

func handleAdminUpdateGame(admin AdminStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := clientStore(r)
		gameID := chi.URLParam(r, "gameID")

		var req AdminGameRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if msg := req.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		scenario, err := admin.GetScenario(r.Context(), req.ScenarioID)
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusBadRequest, "scenario not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		req.ScenarioName = scenario.Name
		req.Mode = scenario.Mode
		req.HasQuestions = scenario.HasQuestions
		if req.Mode == "guided" {
			req.Supervised = true
		}

		game, err := store.UpdateGame(r.Context(), gameID, req, scenario.Stages)
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		teams, err := store.ListTeams(r.Context(), gameID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		game.Teams = teams

		writeJSON(w, http.StatusOK, game)
	}
}

func handleAdminDeleteGame() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := clientStore(r)
		gameID := chi.URLParam(r, "gameID")

		hasPlayers, err := store.GameHasPlayers(r.Context(), gameID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if hasPlayers {
			writeError(w, http.StatusConflict, "cannot delete game with existing players")
			return
		}

		if err := store.DeleteTeamsByGame(r.Context(), gameID); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		if err := store.DeleteGame(r.Context(), gameID); err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusNotFound, "game not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func handleAdminListTeams() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := clientStore(r)
		gameID := chi.URLParam(r, "gameID")

		exists, err := store.GameExists(r.Context(), gameID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if !exists {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}

		teams, err := store.ListTeams(r.Context(), gameID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, teams)
	}
}

func handleAdminCreateTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := clientStore(r)
		gameID := chi.URLParam(r, "gameID")

		var req AdminTeamRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if msg := req.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		exists, err := store.GameExists(r.Context(), gameID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if !exists {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}

		token := req.JoinToken
		if token == "" {
			token = generateJoinToken()
		}

		team, err := store.CreateTeam(r.Context(), gameID, req, token)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				writeError(w, http.StatusConflict, fmt.Sprintf("join token %q already exists", token))
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusCreated, team)
	}
}

func handleAdminUpdateTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := clientStore(r)
		gameID := chi.URLParam(r, "gameID")
		teamID := chi.URLParam(r, "teamID")

		var req AdminTeamRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if msg := req.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		team, err := store.UpdateTeam(r.Context(), gameID, teamID, req)
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "team not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, team)
	}
}

func handleAdminDeleteTeam() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := clientStore(r)
		gameID := chi.URLParam(r, "gameID")
		teamID := chi.URLParam(r, "teamID")

		hasPlayers, err := store.TeamHasPlayers(r.Context(), gameID, teamID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if hasPlayers {
			writeError(w, http.StatusConflict, "cannot delete team with existing players")
			return
		}

		if err := store.DeleteTeam(r.Context(), gameID, teamID); err != nil {
			if errors.Is(err, ErrNotFound) {
				writeError(w, http.StatusNotFound, "team not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func handleAdminGameStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		store := clientStore(r)
		gameID := chi.URLParam(r, "gameID")

		status, err := store.GameStatus(r.Context(), gameID)
		if errors.Is(err, ErrNotFound) {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, status)
	}
}
