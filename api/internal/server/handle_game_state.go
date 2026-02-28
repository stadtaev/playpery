package server

import (
	"encoding/json"
	"net/http"
	"time"
)

type GameInfo struct {
	Status            string  `json:"status"`
	Mode              string  `json:"mode"`
	HasQuestions      bool    `json:"hasQuestions,omitempty"`
	Supervised        bool    `json:"supervised"`
	TimerEnabled      bool    `json:"timerEnabled"`
	TimerMinutes      int     `json:"timerMinutes"`
	StageTimerMinutes int     `json:"stageTimerMinutes"`
	StartedAt         *string `json:"startedAt"`
	TotalStages       int     `json:"totalStages"`
}

type TeamInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type StageInfo struct {
	StageNumber    int    `json:"stageNumber"`
	Clue           string `json:"clue"`
	Question       string `json:"question,omitempty"`
	Location       string `json:"location"`
	Locked         bool   `json:"locked"`
	LocationNumber int    `json:"locationNumber,omitempty"`
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
	Role            string           `json:"role"`
	TeamSecret      int              `json:"teamSecret,omitempty"`
	CurrentStage    *StageInfo       `json:"currentStage"`
	CompletedStages []CompletedStage `json:"completedStages"`
	Players         []PlayerInfo     `json:"players"`
}

type scenarioStage struct {
	StageNumber    int    `json:"stageNumber"`
	Location       string `json:"location"`
	Clue           string `json:"clue"`
	Question       string `json:"question"`
	CorrectAnswer  string `json:"correctAnswer"`
	UnlockCode     string `json:"unlockCode,omitempty"`
	LocationNumber int    `json:"locationNumber,omitempty"`
}

// modeHasQuestion returns true if the mode supports questions at each stage.
func modeHasQuestion(mode string, hasQuestions bool) bool {
	switch mode {
	case "classic", "qr_quiz":
		return true
	case "guided":
		return hasQuestions
	default:
		return false
	}
}

// modeRequiresUnlock returns true if the mode requires unlocking before the question/completion.
func modeRequiresUnlock(mode string) bool {
	switch mode {
	case "qr_quiz", "qr_hunt", "math_puzzle", "guided":
		return true
	default:
		return false
	}
}

// isStageUnlocked checks if a stage number is in the unlocked list.
func isStageUnlocked(unlockedStages []int, stageNumber int) bool {
	for _, n := range unlockedStages {
		if n == stageNumber {
			return true
		}
	}
	return false
}

func handleGameState() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sess, err := playerFromRequest(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid or missing session token")
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
				data.Status = "ended"
				store.ExpireGame(r.Context(), sess.GameID)
			}
		}

		var stages []scenarioStage
		if err := json.Unmarshal([]byte(data.StagesJSON), &stages); err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		completed, err := store.ListCompletedStages(r.Context(), sess.GameID, sess.TeamID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		currentStageNum := len(completed) + 1
		var currentStage *StageInfo
		if currentStageNum <= len(stages) && data.Status == "active" {
			s := stages[currentStageNum-1]
			si := StageInfo{
				StageNumber: s.StageNumber,
				Clue:        s.Clue,
				Location:    s.Location,
			}

			if modeRequiresUnlock(data.Mode) {
				unlocked := isStageUnlocked(data.UnlockedStages, currentStageNum)
				si.Locked = !unlocked
				if unlocked && modeHasQuestion(data.Mode, data.HasQuestions) {
					si.Question = s.Question
				}
				if data.Mode == "math_puzzle" {
					si.LocationNumber = s.LocationNumber
				}
			} else {
				// classic: always show question, never locked
				si.Question = s.Question
			}

			currentStage = &si
		}

		players, err := store.ListPlayers(r.Context(), sess.GameID, sess.TeamID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		resp := GameStateResponse{
			Role: sess.Role,
			Game: GameInfo{
				Status:            data.Status,
				Mode:              data.Mode,
				HasQuestions:      data.HasQuestions,
				Supervised:        data.Supervised,
				TimerEnabled:      data.TimerEnabled,
				TimerMinutes:      data.TimerMinutes,
				StageTimerMinutes: data.StageTimerMinutes,
				StartedAt:         data.StartedAt,
				TotalStages:       len(stages),
			},
			Team: TeamInfo{
				ID:   sess.TeamID,
				Name: data.TeamName,
			},
			CurrentStage:    currentStage,
			CompletedStages: completed,
			Players:         players,
		}
		if data.Mode == "math_puzzle" {
			resp.TeamSecret = data.TeamSecret
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
