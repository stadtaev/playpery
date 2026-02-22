package server

import (
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

		if data.Status == "active" && data.StartedAt != nil {
			start, _ := time.Parse(time.RFC3339Nano, *data.StartedAt)
			if time.Since(start) > time.Duration(data.TimerMinutes)*time.Minute {
				data.Status = "ended"
				store.ExpireGame(r.Context(), sess.GameID)
			}
		}

		var stages []scenarioStage
		json.Unmarshal([]byte(data.StagesJSON), &stages)

		completed, err := store.ListCompletedStages(r.Context(), sess.GameID, sess.TeamID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		currentStageNum := len(completed) + 1
		var currentStage *StageInfo
		if currentStageNum <= len(stages) && data.Status == "active" {
			s := stages[currentStageNum-1]
			currentStage = &StageInfo{
				StageNumber: s.StageNumber,
				Clue:        s.Clue,
				Question:    s.Question,
				Location:    s.Location,
			}
		}

		players, err := store.ListPlayers(r.Context(), sess.GameID, sess.TeamID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		resp := GameStateResponse{
			Game: GameInfo{
				Status:       data.Status,
				TimerMinutes: data.TimerMinutes,
				StartedAt:    data.StartedAt,
				TotalStages:  len(stages),
			},
			Team: TeamInfo{
				ID:   sess.TeamID,
				Name: data.TeamName,
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
