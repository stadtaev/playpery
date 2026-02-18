package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

type AnswerRequest struct {
	Answer string `json:"answer"`
}

type AnswerResponse struct {
	IsCorrect     bool       `json:"isCorrect"`
	StageNumber   int        `json:"stageNumber"`
	NextStage     *StageInfo `json:"nextStage"`
	GameComplete  bool       `json:"gameComplete"`
	CorrectAnswer string     `json:"correctAnswer,omitempty"` // debug: included on wrong answers
}

func handleAnswer(db *sql.DB, broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := playerFromRequest(r, db)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or missing session token")
			return
		}

		var req AnswerRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		req.Answer = strings.TrimSpace(req.Answer)
		if req.Answer == "" {
			writeError(w, http.StatusBadRequest, "answer is required")
			return
		}

		// Fetch game status + stages.
		var gameStatus string
		var startedAt sql.NullString
		var timerMinutes int
		var stagesJSON string
		err = db.QueryRowContext(r.Context(), `
			SELECT g.status, g.started_at, g.timer_minutes, s.stages
			FROM games g
			JOIN scenarios s ON s.id = g.scenario_id
			WHERE g.id = ?
		`, sess.GameID).Scan(&gameStatus, &startedAt, &timerMinutes, &stagesJSON)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		// Check timer.
		if gameStatus == "active" && startedAt.Valid {
			start, _ := time.Parse(time.RFC3339Nano, startedAt.String)
			if time.Since(start) > time.Duration(timerMinutes)*time.Minute {
				db.ExecContext(r.Context(), `
					UPDATE games SET status = 'ended', ended_at = strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
					WHERE id = ? AND status = 'active'
				`, sess.GameID)
				writeError(w, http.StatusConflict, "game has ended")
				return
			}
		}

		if gameStatus != "active" {
			writeError(w, http.StatusConflict, "game is not active")
			return
		}

		var stages []scenarioStage
		json.Unmarshal([]byte(stagesJSON), &stages)

		// Determine current stage by counting correct answers.
		var correctCount int
		db.QueryRowContext(r.Context(), `
			SELECT COUNT(*) FROM stage_results
			WHERE game_id = ? AND team_id = ? AND is_correct = 1
		`, sess.GameID, sess.TeamID).Scan(&correctCount)

		currentStageNum := correctCount + 1
		if currentStageNum > len(stages) {
			writeError(w, http.StatusConflict, "all stages completed")
			return
		}

		stage := stages[currentStageNum-1]
		isCorrect := strings.EqualFold(
			strings.TrimSpace(req.Answer),
			strings.TrimSpace(stage.CorrectAnswer),
		)

		// Record the answer.
		isCorrectInt := 0
		if isCorrect {
			isCorrectInt = 1
		}
		db.ExecContext(r.Context(), `
			INSERT INTO stage_results (game_id, team_id, stage_number, answer, is_correct, answered_at)
			VALUES (?, ?, ?, ?, ?, strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
		`, sess.GameID, sess.TeamID, currentStageNum, req.Answer, isCorrectInt)

		resp := AnswerResponse{
			IsCorrect:   isCorrect,
			StageNumber: currentStageNum,
		}

		if isCorrect {
			nextStageNum := currentStageNum + 1
			if nextStageNum <= len(stages) {
				s := stages[nextStageNum-1]
				resp.NextStage = &StageInfo{
					StageNumber: s.StageNumber,
					Clue:        s.Clue,
					Question:    s.Question,
					Location:    s.Location,
				}
			} else {
				resp.GameComplete = true
				// Check if game should end (all stages done).
			}

			broker.Publish(sess.TeamID, SSEEvent{
				Type:        "stage_completed",
				StageNumber: currentStageNum,
			})
		} else {
			resp.CorrectAnswer = stage.CorrectAnswer
			broker.Publish(sess.TeamID, SSEEvent{
				Type:        "wrong_answer",
				StageNumber: currentStageNum,
			})
		}

		writeJSON(w, http.StatusOK, resp)
	}
}
