package server

import (
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
	CorrectAnswer string     `json:"correctAnswer,omitempty"`
}

func handleAnswer(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := playerFromRequest(r)
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

		store := clientStore(r)

		data, err := store.GameState(r.Context(), sess.GameID, sess.TeamID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		if data.TimerEnabled && data.Status == "active" && data.StartedAt != nil {
			start, _ := time.Parse(time.RFC3339Nano, *data.StartedAt)
			if time.Since(start) > time.Duration(data.TimerMinutes)*time.Minute {
				store.ExpireGame(r.Context(), sess.GameID)
				writeError(w, http.StatusConflict, "game has ended")
				return
			}
		}

		if data.Status != "active" {
			writeError(w, http.StatusConflict, "game is not active")
			return
		}

		if data.Supervised && sess.Role != "supervisor" {
			writeError(w, http.StatusForbidden, "only the supervisor can submit answers")
			return
		}

		var stages []scenarioStage
		if err := json.Unmarshal([]byte(data.StagesJSON), &stages); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		answeredCount, err := store.CountAnsweredStages(r.Context(), sess.GameID, sess.TeamID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		currentStageNum := answeredCount + 1
		if currentStageNum > len(stages) {
			writeError(w, http.StatusConflict, "all stages completed")
			return
		}

		// Mode guards: reject answer if mode doesn't support questions or stage not unlocked.
		if !modeHasQuestion(data.Mode, data.HasQuestions) {
			writeError(w, http.StatusConflict, "this mode does not use questions")
			return
		}
		if modeRequiresUnlock(data.Mode) && !isStageUnlocked(data.UnlockedStages, currentStageNum) {
			writeError(w, http.StatusConflict, "stage not unlocked")
			return
		}

		stage := stages[currentStageNum-1]
		isCorrect := strings.EqualFold(
			strings.TrimSpace(req.Answer),
			strings.TrimSpace(stage.CorrectAnswer),
		)

		if err := store.RecordAnswer(r.Context(), sess.GameID, sess.TeamID, currentStageNum, req.Answer, isCorrect); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		resp := AnswerResponse{
			IsCorrect:   isCorrect,
			StageNumber: currentStageNum,
		}

		// Both correct and incorrect answers advance to the next stage.
		nextStageNum := currentStageNum + 1
		if nextStageNum <= len(stages) {
			s := stages[nextStageNum-1]
			ns := StageInfo{
				StageNumber: s.StageNumber,
				Clue:        s.Clue,
				Location:    s.Location,
				Locked:      modeRequiresUnlock(data.Mode),
			}
			if !ns.Locked {
				ns.Question = s.Question
			}
			resp.NextStage = &ns
		} else {
			resp.GameComplete = true
		}

		if isCorrect {
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
