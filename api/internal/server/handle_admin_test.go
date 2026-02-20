package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func adminRouter(t *testing.T) (*chi.Mux, func() []*http.Cookie) {
	t.Helper()
	db := setupTestDB(t)

	r := chi.NewRouter()
	r.Post("/api/admin/login", handleAdminLogin(db))
	r.Post("/api/admin/logout", handleAdminLogout(db))
	r.Get("/api/admin/me", handleAdminMe(db))
	r.Get("/api/admin/scenarios", handleAdminListScenarios(db))
	r.Post("/api/admin/scenarios", handleAdminCreateScenario(db))
	r.Get("/api/admin/scenarios/{id}", handleAdminGetScenario(db))
	r.Put("/api/admin/scenarios/{id}", handleAdminUpdateScenario(db))
	r.Delete("/api/admin/scenarios/{id}", handleAdminDeleteScenario(db))
	r.Get("/api/admin/games", handleAdminListGames(db))
	r.Post("/api/admin/games", handleAdminCreateGame(db))
	r.Get("/api/admin/games/{gameID}", handleAdminGetGame(db))
	r.Put("/api/admin/games/{gameID}", handleAdminUpdateGame(db))
	r.Delete("/api/admin/games/{gameID}", handleAdminDeleteGame(db))
	r.Get("/api/admin/games/{gameID}/teams", handleAdminListTeams(db))
	r.Post("/api/admin/games/{gameID}/teams", handleAdminCreateTeam(db))
	r.Put("/api/admin/games/{gameID}/teams/{teamID}", handleAdminUpdateTeam(db))
	r.Delete("/api/admin/games/{gameID}/teams/{teamID}", handleAdminDeleteTeam(db))

	// Also register the join endpoint for player tests.
	broker := NewBroker()
	r.Post("/api/join", handleJoin(db, broker))

	// Login helper that returns cookies.
	login := func() []*http.Cookie {
		body, _ := json.Marshal(AdminLoginRequest{Email: "admin@playperu.com", Password: "changeme"})
		req := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewReader(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("login: expected 200, got %d: %s", w.Code, w.Body.String())
		}
		return w.Result().Cookies()
	}

	return r, login
}

func TestAdminLoginGoodCredentials(t *testing.T) {
	r, _ := adminRouter(t)

	body, _ := json.Marshal(AdminLoginRequest{Email: "admin@playperu.com", Password: "changeme"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp AdminMeResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Email != "admin@playperu.com" {
		t.Errorf("expected email admin@playperu.com, got %q", resp.Email)
	}

	// Should have set cookie.
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "admin_session" && c.Value != "" {
			found = true
		}
	}
	if !found {
		t.Error("expected admin_session cookie to be set")
	}
}

func TestAdminLoginBadCredentials(t *testing.T) {
	r, _ := adminRouter(t)

	body, _ := json.Marshal(AdminLoginRequest{Email: "admin@playperu.com", Password: "wrong"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAdminLoginBadEmail(t *testing.T) {
	r, _ := adminRouter(t)

	body, _ := json.Marshal(AdminLoginRequest{Email: "nobody@example.com", Password: "changeme"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAdminMeAuthenticated(t *testing.T) {
	r, login := adminRouter(t)
	cookies := login()

	req := httptest.NewRequest(http.MethodGet, "/api/admin/me", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp AdminMeResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Email != "admin@playperu.com" {
		t.Errorf("expected email admin@playperu.com, got %q", resp.Email)
	}
}

func TestAdminMeUnauthenticated(t *testing.T) {
	r, _ := adminRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAdminLogout(t *testing.T) {
	r, login := adminRouter(t)
	cookies := login()

	// Logout.
	req := httptest.NewRequest(http.MethodPost, "/api/admin/logout", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("logout: expected 200, got %d", w.Code)
	}

	// Session should be invalid now.
	req = httptest.NewRequest(http.MethodGet, "/api/admin/me", nil)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 after logout, got %d", w.Code)
	}
}

func TestAdminScenarioCRUD(t *testing.T) {
	r, login := adminRouter(t)
	cookies := login()

	addCookies := func(req *http.Request) {
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}

	// List scenarios — should have the seeded one.
	req := httptest.NewRequest(http.MethodGet, "/api/admin/scenarios", nil)
	addCookies(req)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var list []AdminScenarioSummary
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) < 1 {
		t.Fatal("list: expected at least 1 scenario (seeded)")
	}

	// Create a new scenario.
	createReq := AdminScenarioRequest{
		Name: "Test Scenario",
		City: "Cusco",
		Stages: []AdminStage{
			{Location: "Plaza de Armas", Clue: "Go to the main square", Question: "What year?", CorrectAnswer: "1534", Lat: -13.516, Lng: -71.978},
			{Location: "Sacsayhuaman", Clue: "Walk uphill", Question: "How many walls?", CorrectAnswer: "3", Lat: -13.509, Lng: -71.982},
		},
	}
	body, _ := json.Marshal(createReq)
	req = httptest.NewRequest(http.MethodPost, "/api/admin/scenarios", bytes.NewReader(body))
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var created AdminScenarioDetail
	json.NewDecoder(w.Body).Decode(&created)
	if created.ID == "" {
		t.Fatal("create: expected non-empty ID")
	}
	if created.Name != "Test Scenario" {
		t.Errorf("create: expected name 'Test Scenario', got %q", created.Name)
	}
	if len(created.Stages) != 2 {
		t.Fatalf("create: expected 2 stages, got %d", len(created.Stages))
	}
	if created.Stages[0].StageNumber != 1 || created.Stages[1].StageNumber != 2 {
		t.Error("create: stage numbers should be normalized to 1, 2")
	}

	// Get by ID.
	req = httptest.NewRequest(http.MethodGet, "/api/admin/scenarios/"+created.ID, nil)
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got AdminScenarioDetail
	json.NewDecoder(w.Body).Decode(&got)
	if got.Name != "Test Scenario" {
		t.Errorf("get: expected name 'Test Scenario', got %q", got.Name)
	}

	// Update.
	updateReq := AdminScenarioRequest{
		Name:        "Updated Scenario",
		City:        "Cusco",
		Description: "Updated description",
		Stages: []AdminStage{
			{Location: "Plaza de Armas", Clue: "Go to the main square", Question: "What year?", CorrectAnswer: "1534"},
		},
	}
	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/admin/scenarios/"+created.ID, bytes.NewReader(body))
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated AdminScenarioDetail
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Name != "Updated Scenario" {
		t.Errorf("update: expected name 'Updated Scenario', got %q", updated.Name)
	}
	if len(updated.Stages) != 1 {
		t.Errorf("update: expected 1 stage, got %d", len(updated.Stages))
	}

	// Delete — should succeed (no games reference it).
	req = httptest.NewRequest(http.MethodDelete, "/api/admin/scenarios/"+created.ID, nil)
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it's gone.
	req = httptest.NewRequest(http.MethodGet, "/api/admin/scenarios/"+created.ID, nil)
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("after delete: expected 404, got %d", w.Code)
	}
}

func TestAdminDeleteScenarioWithGames(t *testing.T) {
	r, login := adminRouter(t)
	cookies := login()

	addCookies := func(req *http.Request) {
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}

	// The seeded scenario has a game referencing it. Get its ID.
	req := httptest.NewRequest(http.MethodGet, "/api/admin/scenarios", nil)
	addCookies(req)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var list []AdminScenarioSummary
	json.NewDecoder(w.Body).Decode(&list)

	// Find the seeded scenario (Lima Centro Historico).
	var seededID string
	for _, s := range list {
		if s.Name == "Lima Centro Historico" {
			seededID = s.ID
			break
		}
	}
	if seededID == "" {
		t.Fatal("could not find seeded scenario")
	}

	// Try to delete — should fail with 409.
	req = httptest.NewRequest(http.MethodDelete, "/api/admin/scenarios/"+seededID, nil)
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminScenariosUnauthenticated(t *testing.T) {
	r, _ := adminRouter(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/admin/scenarios"},
		{http.MethodPost, "/api/admin/scenarios"},
		{http.MethodGet, "/api/admin/scenarios/someid"},
		{http.MethodPut, "/api/admin/scenarios/someid"},
		{http.MethodDelete, "/api/admin/scenarios/someid"},
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(ep.method, ep.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", ep.method, ep.path, w.Code)
		}
	}
}

func TestAdminGameCRUD(t *testing.T) {
	r, login := adminRouter(t)
	cookies := login()

	addCookies := func(req *http.Request) {
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}

	// List games — should have the seeded one.
	req := httptest.NewRequest(http.MethodGet, "/api/admin/games", nil)
	addCookies(req)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var list []AdminGameSummary
	json.NewDecoder(w.Body).Decode(&list)
	if len(list) < 1 {
		t.Fatal("list: expected at least 1 game (seeded)")
	}
	if list[0].ScenarioName == "" {
		t.Error("list: expected scenario name to be populated")
	}
	if list[0].TeamCount < 2 {
		t.Errorf("list: expected at least 2 teams, got %d", list[0].TeamCount)
	}

	// Create a new scenario first (needed for creating a game).
	scenarioReq := AdminScenarioRequest{
		Name: "Game Test Scenario",
		City: "Arequipa",
		Stages: []AdminStage{
			{Location: "Plaza", Question: "What?", CorrectAnswer: "Yes"},
		},
	}
	body, _ := json.Marshal(scenarioReq)
	req = httptest.NewRequest(http.MethodPost, "/api/admin/scenarios", bytes.NewReader(body))
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("create scenario: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var scenario AdminScenarioDetail
	json.NewDecoder(w.Body).Decode(&scenario)

	// Create a game.
	gameReq := AdminGameRequest{
		ScenarioID:   scenario.ID,
		Status:       "draft",
		TimerMinutes: 90,
	}
	body, _ = json.Marshal(gameReq)
	req = httptest.NewRequest(http.MethodPost, "/api/admin/games", bytes.NewReader(body))
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create game: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var game AdminGameDetail
	json.NewDecoder(w.Body).Decode(&game)
	if game.ID == "" {
		t.Fatal("create game: expected non-empty ID")
	}
	if game.Status != "draft" {
		t.Errorf("create game: expected status 'draft', got %q", game.Status)
	}
	if game.TimerMinutes != 90 {
		t.Errorf("create game: expected 90 minutes, got %d", game.TimerMinutes)
	}
	if game.ScenarioName != "Game Test Scenario" {
		t.Errorf("create game: expected scenario name 'Game Test Scenario', got %q", game.ScenarioName)
	}

	// Get game by ID.
	req = httptest.NewRequest(http.MethodGet, "/api/admin/games/"+game.ID, nil)
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get game: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var got AdminGameDetail
	json.NewDecoder(w.Body).Decode(&got)
	if got.Status != "draft" {
		t.Errorf("get game: expected status 'draft', got %q", got.Status)
	}
	if len(got.Teams) != 0 {
		t.Errorf("get game: expected 0 teams, got %d", len(got.Teams))
	}

	// Update game.
	gameReq.Status = "active"
	gameReq.TimerMinutes = 60
	body, _ = json.Marshal(gameReq)
	req = httptest.NewRequest(http.MethodPut, "/api/admin/games/"+game.ID, bytes.NewReader(body))
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update game: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var updated AdminGameDetail
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Status != "active" {
		t.Errorf("update game: expected status 'active', got %q", updated.Status)
	}
	if updated.TimerMinutes != 60 {
		t.Errorf("update game: expected 60 minutes, got %d", updated.TimerMinutes)
	}

	// Add a team.
	teamReq := AdminTeamRequest{Name: "Los Alpacas", GuideName: "Pedro"}
	body, _ = json.Marshal(teamReq)
	req = httptest.NewRequest(http.MethodPost, "/api/admin/games/"+game.ID+"/teams", bytes.NewReader(body))
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create team: expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var team AdminTeamItem
	json.NewDecoder(w.Body).Decode(&team)
	if team.Name != "Los Alpacas" {
		t.Errorf("create team: expected name 'Los Alpacas', got %q", team.Name)
	}
	if team.JoinToken == "" {
		t.Error("create team: expected auto-generated join token")
	}
	if team.GuideName != "Pedro" {
		t.Errorf("create team: expected guide name 'Pedro', got %q", team.GuideName)
	}

	// List teams.
	req = httptest.NewRequest(http.MethodGet, "/api/admin/games/"+game.ID+"/teams", nil)
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("list teams: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var teams []AdminTeamItem
	json.NewDecoder(w.Body).Decode(&teams)
	if len(teams) != 1 {
		t.Fatalf("list teams: expected 1 team, got %d", len(teams))
	}

	// Update team.
	teamReq.Name = "Los Alpacas Updated"
	teamReq.GuideName = "Pedro Jr"
	body, _ = json.Marshal(teamReq)
	req = httptest.NewRequest(http.MethodPut, "/api/admin/games/"+game.ID+"/teams/"+team.ID, bytes.NewReader(body))
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update team: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var updatedTeam AdminTeamItem
	json.NewDecoder(w.Body).Decode(&updatedTeam)
	if updatedTeam.Name != "Los Alpacas Updated" {
		t.Errorf("update team: expected name 'Los Alpacas Updated', got %q", updatedTeam.Name)
	}

	// Delete team (no players, should succeed).
	req = httptest.NewRequest(http.MethodDelete, "/api/admin/games/"+game.ID+"/teams/"+team.ID, nil)
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("delete team: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Delete game (no players, should succeed).
	req = httptest.NewRequest(http.MethodDelete, "/api/admin/games/"+game.ID, nil)
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("delete game: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it's gone.
	req = httptest.NewRequest(http.MethodGet, "/api/admin/games/"+game.ID, nil)
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("after delete: expected 404, got %d", w.Code)
	}
}

func TestAdminDeleteGameWithPlayers(t *testing.T) {
	r, login := adminRouter(t)
	cookies := login()

	addCookies := func(req *http.Request) {
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}

	// Join a player to the seeded team.
	joinBody, _ := json.Marshal(JoinRequest{JoinToken: "incas-2025", PlayerName: "TestPlayer"})
	req := httptest.NewRequest(http.MethodPost, "/api/join", bytes.NewReader(joinBody))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("join: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Try to delete the seeded game — should fail with 409.
	req = httptest.NewRequest(http.MethodDelete, "/api/admin/games/g0000000deadbeef", nil)
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminDeleteTeamWithPlayers(t *testing.T) {
	r, login := adminRouter(t)
	cookies := login()

	addCookies := func(req *http.Request) {
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}

	// Join a player to the seeded team.
	joinBody, _ := json.Marshal(JoinRequest{JoinToken: "condores-2025", PlayerName: "TestPlayer2"})
	req := httptest.NewRequest(http.MethodPost, "/api/join", bytes.NewReader(joinBody))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("join: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Try to delete the team — should fail with 409.
	req = httptest.NewRequest(http.MethodDelete, "/api/admin/games/g0000000deadbeef/teams/t00000000condor", nil)
	addCookies(req)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminCreateTeamDuplicateToken(t *testing.T) {
	r, login := adminRouter(t)
	cookies := login()

	addCookies := func(req *http.Request) {
		for _, c := range cookies {
			req.AddCookie(c)
		}
	}

	// Create a team with a custom token that already exists (incas-2025).
	teamReq := AdminTeamRequest{Name: "Duplicate Team", JoinToken: "incas-2025"}
	body, _ := json.Marshal(teamReq)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/games/g0000000deadbeef/teams", bytes.NewReader(body))
	addCookies(req)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminGamesUnauthenticated(t *testing.T) {
	r, _ := adminRouter(t)

	endpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/admin/games"},
		{http.MethodPost, "/api/admin/games"},
		{http.MethodGet, "/api/admin/games/someid"},
		{http.MethodPut, "/api/admin/games/someid"},
		{http.MethodDelete, "/api/admin/games/someid"},
		{http.MethodGet, "/api/admin/games/someid/teams"},
		{http.MethodPost, "/api/admin/games/someid/teams"},
		{http.MethodPut, "/api/admin/games/someid/teams/someteam"},
		{http.MethodDelete, "/api/admin/games/someid/teams/someteam"},
	}

	for _, ep := range endpoints {
		req := httptest.NewRequest(ep.method, ep.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", ep.method, ep.path, w.Code)
		}
	}
}
