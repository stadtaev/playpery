package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/playperu/cityquiz/internal/database"
)

func setupStores(t *testing.T) (*AdminStore, *DocStore) {
	t.Helper()
	ctx := context.Background()

	// Admin store.
	adminDB, err := database.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open admin db: %v", err)
	}
	admin, err := NewAdminStore(ctx, adminDB)
	if err != nil {
		t.Fatalf("init admin store: %v", err)
	}
	t.Cleanup(func() { adminDB.Close() })

	// Seed demo scenario into admin DB.
	sc, err := admin.SeedDemoScenario(ctx)
	if err != nil {
		t.Fatalf("seed demo scenario: %v", err)
	}

	// Client store.
	clientDB, err := database.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open client db: %v", err)
	}
	store, err := NewDocStore(ctx, clientDB)
	if err != nil {
		t.Fatalf("init doc store: %v", err)
	}
	if sc != nil {
		if err := store.SeedDemoGame(ctx, sc); err != nil {
			t.Fatalf("seed demo game: %v", err)
		}
	}
	t.Cleanup(func() { clientDB.Close() })

	return admin, store
}

func playerRouter(t *testing.T) *chi.Mux {
	t.Helper()
	_, store := setupStores(t)
	broker := NewBroker()

	r := chi.NewRouter()
	// Wrap with a middleware that injects the store into context.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxKeyStore, Store(store))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	r.Get("/api/{client}/teams/{joinToken}", handleTeamLookup())
	r.Post("/api/{client}/join", handleJoin(broker))
	r.Get("/api/{client}/game/state", handleGameState())
	r.Post("/api/{client}/game/answer", handleAnswer(broker))
	return r
}

func TestTeamLookup(t *testing.T) {
	r := playerRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/demo/teams/incas-2025", nil)
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
	r := playerRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/demo/teams/nope-1234", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestJoinAndGameState(t *testing.T) {
	r := playerRouter(t)

	// Join the team.
	body, _ := json.Marshal(JoinRequest{JoinToken: "incas-2025", PlayerName: "Maria"})
	req := httptest.NewRequest(http.MethodPost, "/api/demo/join", bytes.NewReader(body))
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
	req = httptest.NewRequest(http.MethodGet, "/api/demo/game/state", nil)
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
	r := playerRouter(t)

	// Join.
	body, _ := json.Marshal(JoinRequest{JoinToken: "condores-2025", PlayerName: "Carlos"})
	req := httptest.NewRequest(http.MethodPost, "/api/demo/join", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var joinResp JoinResponse
	json.NewDecoder(w.Body).Decode(&joinResp)
	token := joinResp.Token

	// Wrong answer.
	body, _ = json.Marshal(AnswerRequest{Answer: "1900"})
	req = httptest.NewRequest(http.MethodPost, "/api/demo/game/answer", bytes.NewReader(body))
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
	req = httptest.NewRequest(http.MethodPost, "/api/demo/game/answer", bytes.NewReader(body))
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
	req = httptest.NewRequest(http.MethodGet, "/api/demo/game/state", nil)
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
	r := playerRouter(t)

	// Join.
	body, _ := json.Marshal(JoinRequest{JoinToken: "incas-2025", PlayerName: "Ana"})
	req := httptest.NewRequest(http.MethodPost, "/api/demo/join", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var joinResp JoinResponse
	json.NewDecoder(w.Body).Decode(&joinResp)
	token := joinResp.Token

	answers := []string{"1651", "catacombs", "San Martin", "17th"}
	for i, ans := range answers {
		body, _ = json.Marshal(AnswerRequest{Answer: ans})
		req = httptest.NewRequest(http.MethodPost, "/api/demo/game/answer", bytes.NewReader(body))
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
	req = httptest.NewRequest(http.MethodGet, "/api/demo/game/state", nil)
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
	r := playerRouter(t)

	// No token.
	req := httptest.NewRequest(http.MethodGet, "/api/demo/game/state", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	// Bad token.
	req = httptest.NewRequest(http.MethodGet, "/api/demo/game/state", nil)
	req.Header.Set("Authorization", "Bearer bogus")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
