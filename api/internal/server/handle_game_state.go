package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

type GameInfo struct {
	Status       string  `json:"status"`
	TimerMinutes int     `json:"timerMinutes"`
	StartedAt    *string `json:"startedAt"`
	TotalStages  int     `json:"totalStages"`
}

type TeamInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type StageInfo struct {
	StageNumber int    `json:"stageNumber"`
	Clue        string `json:"clue"`
	Question    string `json:"question"`
	Location    string `json:"location"`
}

type CompletedStage struct {
	StageNumber int    `json:"stageNumber"`
	IsCorrect   bool   `json:"isCorrect"`
	AnsweredAt  string `json:"answeredAt"`
}

type PlayerInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type GameStateResponse struct {
	Game            GameInfo         `json:"game"`
	Team            TeamInfo         `json:"team"`
	CurrentStage    *StageInfo       `json:"currentStage"`
	CompletedStages []CompletedStage `json:"completedStages"`
	Players         []PlayerInfo     `json:"players"`
}

type scenarioStage struct {
	StageNumber   int    `json:"stageNumber"`
	Location      string `json:"location"`
	Clue          string `json:"clue"`
	Question      string `json:"question"`
	CorrectAnswer string `json:"correctAnswer"`
}

func handleGameState(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := playerFromRequest(r, db)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or missing session token")
			return
		}

		// Fetch game + scenario info.
		var (
			gameStatus   string
			timerMinutes int
			startedAt    sql.NullString
			stagesJSON   string
			teamName     string
		)
		err = db.QueryRowContext(r.Context(), `
			SELECT g.status, g.timer_minutes, g.started_at, s.stages, t.name
			FROM games g
			JOIN scenarios s ON s.id = g.scenario_id
			JOIN teams t ON t.id = ?
			WHERE g.id = ?
		`, sess.TeamID, sess.GameID).Scan(&gameStatus, &timerMinutes, &startedAt, &stagesJSON, &teamName)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		// Check timer expiry (lazy).
		if gameStatus == "active" && startedAt.Valid {
			start, _ := time.Parse(time.RFC3339Nano, startedAt.String)
			if time.Since(start) > time.Duration(timerMinutes)*time.Minute {
				gameStatus = "ended"
				db.ExecContext(r.Context(), `
					UPDATE games SET status = 'ended', ended_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
					WHERE id = ? AND status = 'active'
				`, sess.GameID)
			}
		}

		var stages []scenarioStage
		json.Unmarshal([]byte(stagesJSON), &stages)

		// Fetch completed stages (correct answers only for advancement).
		rows, err := db.QueryContext(r.Context(), `
			SELECT stage_number, is_correct, answered_at
			FROM stage_results
			WHERE game_id = ? AND team_id = ? AND is_correct = 1
			ORDER BY stage_number
		`, sess.GameID, sess.TeamID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		defer rows.Close()

		var completed []CompletedStage
		for rows.Next() {
			var cs CompletedStage
			var isCorrectInt int
			rows.Scan(&cs.StageNumber, &isCorrectInt, &cs.AnsweredAt)
			cs.IsCorrect = isCorrectInt == 1
			completed = append(completed, cs)
		}

		// Current stage = number of correct answers + 1.
		currentStageNum := len(completed) + 1
		var currentStage *StageInfo
		if currentStageNum <= len(stages) && gameStatus == "active" {
			s := stages[currentStageNum-1]
			currentStage = &StageInfo{
				StageNumber: s.StageNumber,
				Clue:        s.Clue,
				Question:    s.Question,
				Location:    s.Location,
			}
		}

		// Fetch players on this team.
		playerRows, err := db.QueryContext(r.Context(), `
			SELECT id, name FROM players WHERE team_id = ? ORDER BY joined_at
		`, sess.TeamID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		defer playerRows.Close()

		var players []PlayerInfo
		for playerRows.Next() {
			var p PlayerInfo
			playerRows.Scan(&p.ID, &p.Name)
			players = append(players, p)
		}

		var startedAtPtr *string
		if startedAt.Valid {
			startedAtPtr = &startedAt.String
		}

		resp := GameStateResponse{
			Game: GameInfo{
				Status:       gameStatus,
				TimerMinutes: timerMinutes,
				StartedAt:    startedAtPtr,
				TotalStages:  len(stages),
			},
			Team: TeamInfo{
				ID:   sess.TeamID,
				Name: teamName,
			},
			CurrentStage:    currentStage,
			CompletedStages: completed,
			Players:         players,
		}
		if resp.CompletedStages == nil {
			resp.CompletedStages = []CompletedStage{}
		}
		if resp.Players == nil {
			resp.Players = []PlayerInfo{}
		}

		writeJSON(w, http.StatusOK, resp)
	}
}
