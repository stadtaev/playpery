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

// supervisedRouter sets up a chi router with a supervised game (2 stages, 1 team).
// Returns the router, the player join token, and the supervisor join token.
func supervisedRouter(t *testing.T) (*chi.Mux, string, string) {
	t.Helper()
	ctx := context.Background()

	adminDB, err := database.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open admin db: %v", err)
	}
	admin, err := NewAdminDocStore(ctx, adminDB)
	if err != nil {
		t.Fatalf("init admin store: %v", err)
	}
	t.Cleanup(func() { adminDB.Close() })

	clientDB, err := database.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("open client db: %v", err)
	}
	store, err := NewDocStore(ctx, clientDB)
	if err != nil {
		t.Fatalf("init doc store: %v", err)
	}
	t.Cleanup(func() { clientDB.Close() })

	// Create a supervised scenario in the admin DB.
	sc, err := admin.CreateScenario(ctx, AdminScenarioRequest{
		Name: "Supervised Test",
		City: "Lima",
		Mode: "supervised",
		Stages: []AdminStage{
			{StageNumber: 1, Location: "Plaza A", Clue: "Go to A", Question: "What is 1+1?", CorrectAnswer: "2"},
			{StageNumber: 2, Location: "Plaza B", Clue: "Go to B", Question: "What is 2+2?", CorrectAnswer: "4"},
		},
	})
	if err != nil {
		t.Fatalf("create scenario: %v", err)
	}

	// Create an active supervised game.
	g, err := store.CreateGame(ctx, AdminGameRequest{
		ScenarioID:   sc.ID,
		ScenarioName: sc.Name,
		Mode:         "supervised",
		Status:       "active",
		Supervised:   true,
	}, sc.Stages)
	if err != nil {
		t.Fatalf("create game: %v", err)
	}

	// Create a team with both join and supervisor tokens.
	team, err := store.CreateTeam(ctx, g.ID, AdminTeamRequest{Name: "Team Alpha"}, "join-alpha")
	if err != nil {
		t.Fatalf("create team: %v", err)
	}

	broker := NewBroker()
	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), ctxKeyStore, Store(store))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Post("/api/{client}/join", handleJoin(broker))
	r.Get("/api/{client}/game/state", handleGameState())
	r.Post("/api/{client}/game/answer", handleAnswer(broker))
	r.Post("/api/{client}/game/unlock", handleUnlock(broker))

	return r, team.JoinToken, team.SupervisorToken
}

func join(t *testing.T, r *chi.Mux, joinToken, name string) JoinResponse {
	t.Helper()
	body, _ := json.Marshal(JoinRequest{JoinToken: joinToken, PlayerName: name})
	req := httptest.NewRequest(http.MethodPost, "/api/demo/join", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("join %q: expected 200, got %d: %s", name, w.Code, w.Body.String())
	}
	var resp JoinResponse
	json.NewDecoder(w.Body).Decode(&resp)
	return resp
}

func gameState(t *testing.T, r *chi.Mux, token string) GameStateResponse {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/api/demo/game/state", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("game state: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp GameStateResponse
	json.NewDecoder(w.Body).Decode(&resp)
	return resp
}

func postJSON(t *testing.T, r *chi.Mux, path, token string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestSupervisedFlowFull(t *testing.T) {
	r, joinToken, superToken := supervisedRouter(t)

	player := join(t, r, joinToken, "Player")
	super := join(t, r, superToken, "Guide")

	if player.Role != "player" {
		t.Fatalf("expected player role, got %q", player.Role)
	}
	if super.Role != "supervisor" {
		t.Fatalf("expected supervisor role, got %q", super.Role)
	}

	// --- Stage 1: locked, no question visible ---

	state := gameState(t, r, player.Token)
	if state.CurrentStage == nil {
		t.Fatal("expected current stage")
	}
	if !state.CurrentStage.Locked {
		t.Error("stage 1 should be locked")
	}
	if state.CurrentStage.Question != "" {
		t.Error("question should not be visible while locked")
	}

	// Player cannot unlock.
	w := postJSON(t, r, "/api/demo/game/unlock", player.Token, UnlockRequest{})
	if w.Code != http.StatusForbidden {
		t.Errorf("player unlock: expected 403, got %d: %s", w.Code, w.Body.String())
	}

	// Player cannot answer locked stage.
	w = postJSON(t, r, "/api/demo/game/answer", player.Token, AnswerRequest{Answer: "2"})
	if w.Code != http.StatusForbidden {
		t.Errorf("player answer on locked stage: expected 403, got %d: %s", w.Code, w.Body.String())
	}

	// Supervisor unlocks stage 1.
	w = postJSON(t, r, "/api/demo/game/unlock", super.Token, UnlockRequest{})
	if w.Code != http.StatusOK {
		t.Fatalf("supervisor unlock: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var unlockResp UnlockResponse
	json.NewDecoder(w.Body).Decode(&unlockResp)
	if !unlockResp.Unlocked {
		t.Error("expected unlocked=true")
	}
	if unlockResp.Question != "What is 1+1?" {
		t.Errorf("expected question after unlock, got %q", unlockResp.Question)
	}

	// After unlock: stage is unlocked, question visible.
	state = gameState(t, r, player.Token)
	if state.CurrentStage.Locked {
		t.Error("stage 1 should be unlocked after supervisor unlock")
	}
	if state.CurrentStage.Question != "What is 1+1?" {
		t.Errorf("expected question visible, got %q", state.CurrentStage.Question)
	}

	// Player still cannot answer (supervised: only supervisor answers).
	w = postJSON(t, r, "/api/demo/game/answer", player.Token, AnswerRequest{Answer: "2"})
	if w.Code != http.StatusForbidden {
		t.Errorf("player answer after unlock: expected 403, got %d: %s", w.Code, w.Body.String())
	}

	// Supervisor answers stage 1.
	w = postJSON(t, r, "/api/demo/game/answer", super.Token, AnswerRequest{Answer: "2"})
	if w.Code != http.StatusOK {
		t.Fatalf("supervisor answer stage 1: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var ansResp AnswerResponse
	json.NewDecoder(w.Body).Decode(&ansResp)
	if !ansResp.IsCorrect {
		t.Error("expected correct answer")
	}
	if ansResp.NextStage == nil {
		t.Fatal("expected next stage")
	}
	if !ansResp.NextStage.Locked {
		t.Error("next stage should be locked")
	}

	// --- Stage 2: locked, player role must still be "player" ---

	state = gameState(t, r, player.Token)
	if state.Role != "player" {
		t.Errorf("player role after stage 1: expected 'player', got %q", state.Role)
	}
	if state.CurrentStage == nil {
		t.Fatal("expected current stage 2")
	}
	if state.CurrentStage.StageNumber != 2 {
		t.Errorf("expected stage 2, got %d", state.CurrentStage.StageNumber)
	}
	if !state.CurrentStage.Locked {
		t.Error("stage 2 should be locked")
	}

	// Player cannot unlock stage 2.
	w = postJSON(t, r, "/api/demo/game/unlock", player.Token, UnlockRequest{})
	if w.Code != http.StatusForbidden {
		t.Errorf("player unlock stage 2: expected 403, got %d: %s", w.Code, w.Body.String())
	}

	// Supervisor unlocks stage 2.
	w = postJSON(t, r, "/api/demo/game/unlock", super.Token, UnlockRequest{})
	if w.Code != http.StatusOK {
		t.Fatalf("supervisor unlock stage 2: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Supervisor answers stage 2 (completes game).
	w = postJSON(t, r, "/api/demo/game/answer", super.Token, AnswerRequest{Answer: "4"})
	if w.Code != http.StatusOK {
		t.Fatalf("supervisor answer stage 2: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	json.NewDecoder(w.Body).Decode(&ansResp)
	if !ansResp.IsCorrect {
		t.Error("expected correct answer for stage 2")
	}
	if !ansResp.GameComplete {
		t.Error("expected game complete after stage 2")
	}

	// Final state: no current stage, 2 completed.
	state = gameState(t, r, player.Token)
	if state.CurrentStage != nil {
		t.Error("expected no current stage after game complete")
	}
	if len(state.CompletedStages) != 2 {
		t.Errorf("expected 2 completed stages, got %d", len(state.CompletedStages))
	}
}
