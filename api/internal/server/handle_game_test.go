package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/playperu/cityquiz/internal/database"
	"github.com/playperu/cityquiz/internal/migrations"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := migrations.Run(db); err != nil {
		t.Fatalf("run migrations: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestTeamLookup(t *testing.T) {
	db := setupTestDB(t)

	r := chi.NewRouter()
	r.Get("/api/teams/{joinToken}", handleTeamLookup(db))

	req := httptest.NewRequest(http.MethodGet, "/api/teams/incas-2025", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp TeamLookupResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Name != "Los Incas" {
		t.Errorf("expected team name 'Los Incas', got %q", resp.Name)
	}
	if resp.GameName != "Lima Centro Historico" {
		t.Errorf("expected game name 'Lima Centro Historico', got %q", resp.GameName)
	}
}

func TestTeamLookupNotFound(t *testing.T) {
	db := setupTestDB(t)

	r := chi.NewRouter()
	r.Get("/api/teams/{joinToken}", handleTeamLookup(db))

	req := httptest.NewRequest(http.MethodGet, "/api/teams/nope-1234", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestJoinAndGameState(t *testing.T) {
	db := setupTestDB(t)
	broker := NewBroker()

	r := chi.NewRouter()
	r.Post("/api/join", handleJoin(db, broker))
	r.Get("/api/game/state", handleGameState(db))

	// Join the team.
	body, _ := json.Marshal(JoinRequest{JoinToken: "incas-2025", PlayerName: "Maria"})
	req := httptest.NewRequest(http.MethodPost, "/api/join", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("join: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var joinResp JoinResponse
	json.NewDecoder(w.Body).Decode(&joinResp)

	if joinResp.Token == "" {
		t.Fatal("join: expected a session token")
	}
	if joinResp.TeamName != "Los Incas" {
		t.Errorf("join: expected team name 'Los Incas', got %q", joinResp.TeamName)
	}

	// Fetch game state.
	req = httptest.NewRequest(http.MethodGet, "/api/game/state", nil)
	req.Header.Set("Authorization", "Bearer "+joinResp.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("state: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var state GameStateResponse
	json.NewDecoder(w.Body).Decode(&state)

	if state.Game.Status != "active" {
		t.Errorf("state: expected game status 'active', got %q", state.Game.Status)
	}
	if state.Game.TotalStages != 4 {
		t.Errorf("state: expected 4 total stages, got %d", state.Game.TotalStages)
	}
	if state.CurrentStage == nil {
		t.Fatal("state: expected a current stage")
	}
	if state.CurrentStage.StageNumber != 1 {
		t.Errorf("state: expected stage 1, got %d", state.CurrentStage.StageNumber)
	}
	if len(state.Players) != 1 || state.Players[0].Name != "Maria" {
		t.Errorf("state: expected 1 player named Maria, got %v", state.Players)
	}
}

func TestAnswerFlow(t *testing.T) {
	db := setupTestDB(t)
	broker := NewBroker()

	r := chi.NewRouter()
	r.Post("/api/join", handleJoin(db, broker))
	r.Get("/api/game/state", handleGameState(db))
	r.Post("/api/game/answer", handleAnswer(db, broker))

	// Join.
	body, _ := json.Marshal(JoinRequest{JoinToken: "condores-2025", PlayerName: "Carlos"})
	req := httptest.NewRequest(http.MethodPost, "/api/join", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var joinResp JoinResponse
	json.NewDecoder(w.Body).Decode(&joinResp)
	token := joinResp.Token

	// Wrong answer.
	body, _ = json.Marshal(AnswerRequest{Answer: "1900"})
	req = httptest.NewRequest(http.MethodPost, "/api/game/answer", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("wrong answer: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var ansResp AnswerResponse
	json.NewDecoder(w.Body).Decode(&ansResp)
	if ansResp.IsCorrect {
		t.Error("wrong answer: expected isCorrect=false")
	}
	if ansResp.StageNumber != 1 {
		t.Errorf("wrong answer: expected stage 1, got %d", ansResp.StageNumber)
	}

	// Correct answer for stage 1.
	body, _ = json.Marshal(AnswerRequest{Answer: "1651"})
	req = httptest.NewRequest(http.MethodPost, "/api/game/answer", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	json.NewDecoder(w.Body).Decode(&ansResp)
	if !ansResp.IsCorrect {
		t.Error("correct answer: expected isCorrect=true")
	}
	if ansResp.NextStage == nil {
		t.Fatal("correct answer: expected nextStage")
	}
	if ansResp.NextStage.StageNumber != 2 {
		t.Errorf("correct answer: expected next stage 2, got %d", ansResp.NextStage.StageNumber)
	}

	// Verify state advances to stage 2.
	req = httptest.NewRequest(http.MethodGet, "/api/game/state", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var state GameStateResponse
	json.NewDecoder(w.Body).Decode(&state)
	if state.CurrentStage.StageNumber != 2 {
		t.Errorf("state after correct: expected stage 2, got %d", state.CurrentStage.StageNumber)
	}
	if len(state.CompletedStages) != 1 {
		t.Errorf("state after correct: expected 1 completed stage, got %d", len(state.CompletedStages))
	}
}

func TestCompleteAllStages(t *testing.T) {
	db := setupTestDB(t)
	broker := NewBroker()

	r := chi.NewRouter()
	r.Post("/api/join", handleJoin(db, broker))
	r.Post("/api/game/answer", handleAnswer(db, broker))
	r.Get("/api/game/state", handleGameState(db))

	// Join.
	body, _ := json.Marshal(JoinRequest{JoinToken: "incas-2025", PlayerName: "Ana"})
	req := httptest.NewRequest(http.MethodPost, "/api/join", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var joinResp JoinResponse
	json.NewDecoder(w.Body).Decode(&joinResp)
	token := joinResp.Token

	answers := []string{"1651", "catacombs", "San Martin", "17th"}
	for i, ans := range answers {
		body, _ = json.Marshal(AnswerRequest{Answer: ans})
		req = httptest.NewRequest(http.MethodPost, "/api/game/answer", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("stage %d: expected 200, got %d: %s", i+1, w.Code, w.Body.String())
		}

		var ansResp AnswerResponse
		json.NewDecoder(w.Body).Decode(&ansResp)
		if !ansResp.IsCorrect {
			t.Fatalf("stage %d: expected correct", i+1)
		}

		if i == len(answers)-1 {
			if !ansResp.GameComplete {
				t.Error("last stage: expected gameComplete=true")
			}
		} else {
			if ansResp.NextStage == nil {
				t.Fatalf("stage %d: expected nextStage", i+1)
			}
		}
	}

	// Verify game state shows no current stage.
	req = httptest.NewRequest(http.MethodGet, "/api/game/state", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var state GameStateResponse
	json.NewDecoder(w.Body).Decode(&state)
	if state.CurrentStage != nil {
		t.Error("after completion: expected no current stage")
	}
	if len(state.CompletedStages) != 4 {
		t.Errorf("after completion: expected 4 completed stages, got %d", len(state.CompletedStages))
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	db := setupTestDB(t)

	r := chi.NewRouter()
	r.Get("/api/game/state", handleGameState(db))

	// No token.
	req := httptest.NewRequest(http.MethodGet, "/api/game/state", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	// Bad token.
	req = httptest.NewRequest(http.MethodGet, "/api/game/state", nil)
	req.Header.Set("Authorization", "Bearer bogus")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
