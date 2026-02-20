package server

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

// AdminGameSummary is returned in the list endpoint.
type AdminGameSummary struct {
	ID           string `json:"id"`
	ScenarioID   string `json:"scenarioId"`
	ScenarioName string `json:"scenarioName"`
	Status       string `json:"status"`
	TimerMinutes int    `json:"timerMinutes"`
	TeamCount    int    `json:"teamCount"`
	CreatedAt    string `json:"createdAt"`
}

// AdminGameDetail is the full game with nested teams.
type AdminGameDetail struct {
	ID           string          `json:"id"`
	ScenarioID   string          `json:"scenarioId"`
	ScenarioName string          `json:"scenarioName"`
	Status       string          `json:"status"`
	TimerMinutes int             `json:"timerMinutes"`
	Teams        []AdminTeamItem `json:"teams"`
	CreatedAt    string          `json:"createdAt"`
}

// AdminTeamItem represents a team within a game.
type AdminTeamItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	JoinToken   string `json:"joinToken"`
	GuideName   string `json:"guideName"`
	PlayerCount int    `json:"playerCount"`
	CreatedAt   string `json:"createdAt"`
}

// AdminGameRequest is the request body for creating/updating a game.
type AdminGameRequest struct {
	ScenarioID   string `json:"scenarioId"`
	Status       string `json:"status"`
	TimerMinutes int    `json:"timerMinutes"`
}

// AdminTeamRequest is the request body for creating/updating a team.
type AdminTeamRequest struct {
	Name      string `json:"name"`
	JoinToken string `json:"joinToken"`
	GuideName string `json:"guideName"`
}

var validGameStatuses = map[string]bool{
	"draft":  true,
	"active": true,
	"paused": true,
	"ended":  true,
}

const demoClientID = "c0000000deadbeef"

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
	if req.TimerMinutes <= 0 {
		req.TimerMinutes = 120
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

func handleAdminListGames(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		rows, err := db.QueryContext(r.Context(), `
			SELECT g.id, g.scenario_id, s.name, g.status, g.timer_minutes,
				(SELECT COUNT(*) FROM teams t WHERE t.game_id = g.id),
				g.created_at
			FROM games g
			JOIN scenarios s ON s.id = g.scenario_id
			ORDER BY g.created_at DESC
		`)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		defer rows.Close()

		var games []AdminGameSummary
		for rows.Next() {
			var g AdminGameSummary
			if err := rows.Scan(&g.ID, &g.ScenarioID, &g.ScenarioName, &g.Status, &g.TimerMinutes, &g.TeamCount, &g.CreatedAt); err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			games = append(games, g)
		}

		if games == nil {
			games = []AdminGameSummary{}
		}
		writeJSON(w, http.StatusOK, games)
	}
}

func handleAdminCreateGame(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		var req AdminGameRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if msg := req.validate(); msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}

		// Verify scenario exists.
		var scenarioName string
		err := db.QueryRowContext(r.Context(), `SELECT name FROM scenarios WHERE id = ?`, req.ScenarioID).Scan(&scenarioName)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "scenario not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		var id, createdAt string
		err = db.QueryRowContext(r.Context(), `
			INSERT INTO games (id, scenario_id, client_id, status, timer_minutes)
			VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?)
			RETURNING id, created_at
		`, req.ScenarioID, demoClientID, req.Status, req.TimerMinutes).Scan(&id, &createdAt)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusCreated, AdminGameDetail{
			ID:           id,
			ScenarioID:   req.ScenarioID,
			ScenarioName: scenarioName,
			Status:       req.Status,
			TimerMinutes: req.TimerMinutes,
			Teams:        []AdminTeamItem{},
			CreatedAt:    createdAt,
		})
	}
}

func handleAdminGetGame(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		gameID := chi.URLParam(r, "gameID")

		var g AdminGameDetail
		err := db.QueryRowContext(r.Context(), `
			SELECT g.id, g.scenario_id, s.name, g.status, g.timer_minutes, g.created_at
			FROM games g
			JOIN scenarios s ON s.id = g.scenario_id
			WHERE g.id = ?
		`, gameID).Scan(&g.ID, &g.ScenarioID, &g.ScenarioName, &g.Status, &g.TimerMinutes, &g.CreatedAt)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		teams, err := queryTeams(r, db, gameID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		g.Teams = teams

		writeJSON(w, http.StatusOK, g)
	}
}

func handleAdminUpdateGame(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

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

		// Verify scenario exists.
		var scenarioName string
		err := db.QueryRowContext(r.Context(), `SELECT name FROM scenarios WHERE id = ?`, req.ScenarioID).Scan(&scenarioName)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusBadRequest, "scenario not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		var createdAt string
		err = db.QueryRowContext(r.Context(), `
			UPDATE games SET scenario_id = ?, status = ?, timer_minutes = ?
			WHERE id = ?
			RETURNING created_at
		`, req.ScenarioID, req.Status, req.TimerMinutes, gameID).Scan(&createdAt)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		teams, err := queryTeams(r, db, gameID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, AdminGameDetail{
			ID:           gameID,
			ScenarioID:   req.ScenarioID,
			ScenarioName: scenarioName,
			Status:       req.Status,
			TimerMinutes: req.TimerMinutes,
			Teams:        teams,
			CreatedAt:    createdAt,
		})
	}
}

func handleAdminDeleteGame(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		gameID := chi.URLParam(r, "gameID")

		// Block if any team has players.
		var playerCount int
		err := db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM players p
			JOIN teams t ON t.id = p.team_id
			WHERE t.game_id = ?
		`, gameID).Scan(&playerCount)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if playerCount > 0 {
			writeError(w, http.StatusConflict, "cannot delete game with existing players")
			return
		}

		// Cascade-delete empty teams, then the game.
		_, err = db.ExecContext(r.Context(), `DELETE FROM teams WHERE game_id = ?`, gameID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		result, err := db.ExecContext(r.Context(), `DELETE FROM games WHERE id = ?`, gameID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func handleAdminListTeams(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		gameID := chi.URLParam(r, "gameID")

		// Verify game exists.
		var exists int
		err := db.QueryRowContext(r.Context(), `SELECT 1 FROM games WHERE id = ?`, gameID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		teams, err := queryTeams(r, db, gameID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, teams)
	}
}

func handleAdminCreateTeam(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

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

		// Verify game exists.
		var exists int
		err := db.QueryRowContext(r.Context(), `SELECT 1 FROM games WHERE id = ?`, gameID).Scan(&exists)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "game not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		token := req.JoinToken
		if token == "" {
			token = generateJoinToken()
		}

		var id, createdAt string
		err = db.QueryRowContext(r.Context(), `
			INSERT INTO teams (id, game_id, name, join_token, guide_name)
			VALUES (lower(hex(randomblob(16))), ?, ?, ?, ?)
			RETURNING id, created_at
		`, gameID, req.Name, token, sql.NullString{String: req.GuideName, Valid: req.GuideName != ""}).Scan(&id, &createdAt)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				writeError(w, http.StatusConflict, fmt.Sprintf("join token %q already exists", token))
				return
			}
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusCreated, AdminTeamItem{
			ID:          id,
			Name:        req.Name,
			JoinToken:   token,
			GuideName:   req.GuideName,
			PlayerCount: 0,
			CreatedAt:   createdAt,
		})
	}
}

func handleAdminUpdateTeam(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

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

		var joinToken, createdAt string
		var playerCount int
		err := db.QueryRowContext(r.Context(), `
			UPDATE teams SET name = ?, guide_name = ?
			WHERE id = ? AND game_id = ?
			RETURNING join_token, created_at, (SELECT COUNT(*) FROM players WHERE team_id = teams.id)
		`, req.Name, sql.NullString{String: req.GuideName, Valid: req.GuideName != ""}, teamID, gameID).Scan(&joinToken, &createdAt, &playerCount)
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "team not found")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		writeJSON(w, http.StatusOK, AdminTeamItem{
			ID:          teamID,
			Name:        req.Name,
			JoinToken:   joinToken,
			GuideName:   req.GuideName,
			PlayerCount: playerCount,
			CreatedAt:   createdAt,
		})
	}
}

func handleAdminDeleteTeam(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := adminFromRequest(r, db); err != nil {
			writeError(w, http.StatusUnauthorized, "not authenticated")
			return
		}

		gameID := chi.URLParam(r, "gameID")
		teamID := chi.URLParam(r, "teamID")

		// Block if players exist.
		var playerCount int
		err := db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM players WHERE team_id = ? AND ? IN (SELECT game_id FROM teams WHERE id = ?)
		`, teamID, gameID, teamID).Scan(&playerCount)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if playerCount > 0 {
			writeError(w, http.StatusConflict, "cannot delete team with existing players")
			return
		}

		result, err := db.ExecContext(r.Context(), `DELETE FROM teams WHERE id = ? AND game_id = ?`, teamID, gameID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		rows, _ := result.RowsAffected()
		if rows == 0 {
			writeError(w, http.StatusNotFound, "team not found")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// queryTeams returns all teams for a game with player counts.
func queryTeams(r *http.Request, db *sql.DB, gameID string) ([]AdminTeamItem, error) {
	rows, err := db.QueryContext(r.Context(), `
		SELECT t.id, t.name, t.join_token, COALESCE(t.guide_name, ''),
			(SELECT COUNT(*) FROM players p WHERE p.team_id = t.id),
			t.created_at
		FROM teams t
		WHERE t.game_id = ?
		ORDER BY t.created_at
	`, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	teams := []AdminTeamItem{}
	for rows.Next() {
		var t AdminTeamItem
		if err := rows.Scan(&t.ID, &t.Name, &t.JoinToken, &t.GuideName, &t.PlayerCount, &t.CreatedAt); err != nil {
			return nil, err
		}
		teams = append(teams, t)
	}
	return teams, nil
}
