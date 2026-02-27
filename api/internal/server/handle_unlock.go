package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type UnlockRequest struct {
	Code string `json:"code"`
}

type UnlockResponse struct {
	StageNumber   int        `json:"stageNumber"`
	Unlocked      bool       `json:"unlocked"`
	StageComplete bool       `json:"stageComplete,omitempty"`
	NextStage     *StageInfo `json:"nextStage,omitempty"`
	GameComplete  bool       `json:"gameComplete,omitempty"`
	Question      string     `json:"question,omitempty"`
}

func handleUnlock(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := playerFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or missing session token")
			return
		}

		var req UnlockRequest
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		req.Code = strings.TrimSpace(req.Code)

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

		if data.Mode == "classic" {
			writeError(w, http.StatusConflict, "classic mode does not use unlock")
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

		if isStageUnlocked(data.UnlockedStages, currentStageNum) {
			writeError(w, http.StatusConflict, "stage already unlocked")
			return
		}

		stage := stages[currentStageNum-1]

		switch data.Mode {
		case "qr_quiz":
			if req.Code == "" {
				writeError(w, http.StatusBadRequest, "code is required")
				return
			}
			if !strings.EqualFold(req.Code, stage.UnlockCode) {
				writeError(w, http.StatusUnprocessableEntity, "invalid code")
				return
			}
			if err := store.UnlockStage(r.Context(), sess.GameID, sess.TeamID, currentStageNum); err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			broker.Publish(sess.TeamID, SSEEvent{
				Type:        "stage_unlocked",
				StageNumber: currentStageNum,
			})
			writeJSON(w, http.StatusOK, UnlockResponse{
				StageNumber: currentStageNum,
				Unlocked:    true,
				Question:    stage.Question,
			})

		case "qr_hunt":
			if req.Code == "" {
				writeError(w, http.StatusBadRequest, "code is required")
				return
			}
			if !strings.EqualFold(req.Code, stage.UnlockCode) {
				writeError(w, http.StatusUnprocessableEntity, "invalid code")
				return
			}
			if err := store.UnlockStage(r.Context(), sess.GameID, sess.TeamID, currentStageNum); err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			// Auto-complete: no question in qr_hunt.
			if err := store.RecordAnswer(r.Context(), sess.GameID, sess.TeamID, currentStageNum, "", true); err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			resp := UnlockResponse{
				StageNumber:   currentStageNum,
				Unlocked:      true,
				StageComplete: true,
			}
			nextStageNum := currentStageNum + 1
			if nextStageNum <= len(stages) {
				s := stages[nextStageNum-1]
				resp.NextStage = &StageInfo{
					StageNumber: s.StageNumber,
					Clue:        s.Clue,
					Location:    s.Location,
					Locked:      true,
				}
			} else {
				resp.GameComplete = true
			}
			broker.Publish(sess.TeamID, SSEEvent{
				Type:        "stage_completed",
				StageNumber: currentStageNum,
			})
			writeJSON(w, http.StatusOK, resp)

		case "math_puzzle":
			if req.Code == "" {
				writeError(w, http.StatusBadRequest, "code is required")
				return
			}
			expected := strconv.Itoa(data.TeamSecret + stage.LocationNumber)
			if req.Code != expected {
				writeError(w, http.StatusUnprocessableEntity, "invalid code")
				return
			}
			if err := store.UnlockStage(r.Context(), sess.GameID, sess.TeamID, currentStageNum); err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			// Auto-complete: no question in math_puzzle.
			if err := store.RecordAnswer(r.Context(), sess.GameID, sess.TeamID, currentStageNum, "", true); err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			resp := UnlockResponse{
				StageNumber:   currentStageNum,
				Unlocked:      true,
				StageComplete: true,
			}
			nextStageNum := currentStageNum + 1
			if nextStageNum <= len(stages) {
				s := stages[nextStageNum-1]
				resp.NextStage = &StageInfo{
					StageNumber:    s.StageNumber,
					Clue:           s.Clue,
					Location:       s.Location,
					Locked:         true,
					LocationNumber: s.LocationNumber,
				}
			} else {
				resp.GameComplete = true
			}
			broker.Publish(sess.TeamID, SSEEvent{
				Type:        "stage_completed",
				StageNumber: currentStageNum,
			})
			writeJSON(w, http.StatusOK, resp)

		case "guided":
			if sess.Role != "supervisor" {
				writeError(w, http.StatusForbidden, "only the supervisor can unlock stages")
				return
			}
			if err := store.UnlockStage(r.Context(), sess.GameID, sess.TeamID, currentStageNum); err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
			if data.HasQuestions {
				broker.Publish(sess.TeamID, SSEEvent{
					Type:        "stage_unlocked",
					StageNumber: currentStageNum,
				})
				writeJSON(w, http.StatusOK, UnlockResponse{
					StageNumber: currentStageNum,
					Unlocked:    true,
					Question:    stage.Question,
				})
			} else {
				// No questions — auto-complete.
				if err := store.RecordAnswer(r.Context(), sess.GameID, sess.TeamID, currentStageNum, "", true); err != nil {
					writeError(w, http.StatusInternalServerError, "internal error")
					return
				}
				resp := UnlockResponse{
					StageNumber:   currentStageNum,
					Unlocked:      true,
					StageComplete: true,
				}
				nextStageNum := currentStageNum + 1
				if nextStageNum <= len(stages) {
					s := stages[nextStageNum-1]
					resp.NextStage = &StageInfo{
						StageNumber: s.StageNumber,
						Clue:        s.Clue,
						Location:    s.Location,
						Locked:      true,
					}
				} else {
					resp.GameComplete = true
				}
				broker.Publish(sess.TeamID, SSEEvent{
					Type:        "stage_completed",
					StageNumber: currentStageNum,
				})
				writeJSON(w, http.StatusOK, resp)
			}

		default:
			writeError(w, http.StatusConflict, "unknown mode")
		}
	}
}
