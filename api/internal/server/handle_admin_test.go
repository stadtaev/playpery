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
