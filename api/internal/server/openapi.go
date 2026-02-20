package server

import (
	"encoding/json"
	"net/http"

	openapi "github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

// ErrorResponse is returned for all error responses.
type ErrorResponse struct {
	Error string `json:"error"`
}

func newOpenAPISpec() *openapi3.Spec {
	r := openapi3.NewReflector()
	r.Spec.Info.Title = "CityQuiz API"
	r.Spec.Info.Version = "0.1.0"
	r.Spec.Info.WithDescription("Backend API for the CityQuiz game.")

	// GET /healthz
	getHealthz, _ := r.NewOperationContext(http.MethodGet, "/healthz")
	getHealthz.SetSummary("Health check")
	getHealthz.SetDescription("Returns the health status of backend dependencies.")
	getHealthz.AddRespStructure(HealthResponse{}, openapi.WithHTTPStatus(http.StatusOK))
	getHealthz.AddRespStructure(HealthResponse{}, openapi.WithHTTPStatus(http.StatusServiceUnavailable))
	_ = r.AddOperation(getHealthz)

	// GET /ws/echo
	getWSEcho, _ := r.NewOperationContext(http.MethodGet, "/ws/echo")
	getWSEcho.SetSummary("WebSocket echo")
	getWSEcho.SetDescription("Upgrades to a WebSocket connection that echoes messages back.")
	getWSEcho.AddRespStructure(nil, openapi.WithHTTPStatus(http.StatusSwitchingProtocols),
		openapi.WithContentType("text/plain"))
	_ = r.AddOperation(getWSEcho)

	// GET /api/teams/{joinToken}
	getTeam, _ := r.NewOperationContext(http.MethodGet, "/api/teams/{joinToken}")
	getTeam.SetSummary("Look up team")
	getTeam.SetDescription("Look up a team by its join token before joining.")
	getTeam.AddRespStructure(TeamLookupResponse{}, openapi.WithHTTPStatus(http.StatusOK))
	getTeam.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusNotFound))
	_ = r.AddOperation(getTeam)

	// POST /api/join
	postJoin, _ := r.NewOperationContext(http.MethodPost, "/api/join")
	postJoin.SetSummary("Join a team")
	postJoin.SetDescription("Player joins a team using the join token. Returns a session token.")
	postJoin.AddReqStructure(JoinRequest{})
	postJoin.AddRespStructure(JoinResponse{}, openapi.WithHTTPStatus(http.StatusOK))
	postJoin.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusNotFound))
	postJoin.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusBadRequest))
	_ = r.AddOperation(postJoin)

	// GET /api/game/state
	getState, _ := r.NewOperationContext(http.MethodGet, "/api/game/state")
	getState.SetSummary("Get game state")
	getState.SetDescription("Returns the full game state for the player's team. Requires Bearer token.")
	getState.AddRespStructure(GameStateResponse{}, openapi.WithHTTPStatus(http.StatusOK))
	getState.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(getState)

	// POST /api/game/answer
	postAnswer, _ := r.NewOperationContext(http.MethodPost, "/api/game/answer")
	postAnswer.SetSummary("Submit answer")
	postAnswer.SetDescription("Submit an answer for the current stage. Requires Bearer token.")
	postAnswer.AddReqStructure(AnswerRequest{})
	postAnswer.AddRespStructure(AnswerResponse{}, openapi.WithHTTPStatus(http.StatusOK))
	postAnswer.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	postAnswer.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusConflict))
	_ = r.AddOperation(postAnswer)

	// GET /api/game/events
	getEvents, _ := r.NewOperationContext(http.MethodGet, "/api/game/events")
	getEvents.SetSummary("SSE event stream")
	getEvents.SetDescription("Server-Sent Events stream for real-time game updates. Pass token as query parameter.")
	getEvents.AddRespStructure(nil, openapi.WithHTTPStatus(http.StatusOK),
		openapi.WithContentType("text/event-stream"))
	_ = r.AddOperation(getEvents)

	// POST /api/admin/login
	postLogin, _ := r.NewOperationContext(http.MethodPost, "/api/admin/login")
	postLogin.SetSummary("Admin login")
	postLogin.SetDescription("Authenticate with email and password. Sets admin_session cookie.")
	postLogin.AddReqStructure(AdminLoginRequest{})
	postLogin.AddRespStructure(AdminMeResponse{}, openapi.WithHTTPStatus(http.StatusOK))
	postLogin.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(postLogin)

	// POST /api/admin/logout
	postLogout, _ := r.NewOperationContext(http.MethodPost, "/api/admin/logout")
	postLogout.SetSummary("Admin logout")
	postLogout.SetDescription("Clears admin session and cookie.")
	postLogout.AddRespStructure(nil, openapi.WithHTTPStatus(http.StatusOK))
	_ = r.AddOperation(postLogout)

	// GET /api/admin/me
	getMe, _ := r.NewOperationContext(http.MethodGet, "/api/admin/me")
	getMe.SetSummary("Current admin")
	getMe.SetDescription("Returns the currently authenticated admin. Requires admin_session cookie.")
	getMe.AddRespStructure(AdminMeResponse{}, openapi.WithHTTPStatus(http.StatusOK))
	getMe.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(getMe)

	// GET /api/admin/scenarios
	listScenarios, _ := r.NewOperationContext(http.MethodGet, "/api/admin/scenarios")
	listScenarios.SetSummary("List scenarios")
	listScenarios.SetDescription("Returns all scenarios with stage counts. Requires admin_session cookie.")
	listScenarios.AddRespStructure([]AdminScenarioSummary{}, openapi.WithHTTPStatus(http.StatusOK))
	listScenarios.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(listScenarios)

	// POST /api/admin/scenarios
	createScenario, _ := r.NewOperationContext(http.MethodPost, "/api/admin/scenarios")
	createScenario.SetSummary("Create scenario")
	createScenario.SetDescription("Creates a new scenario with stages. Requires admin_session cookie.")
	createScenario.AddReqStructure(AdminScenarioRequest{})
	createScenario.AddRespStructure(AdminScenarioDetail{}, openapi.WithHTTPStatus(http.StatusCreated))
	createScenario.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusBadRequest))
	createScenario.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(createScenario)

	// GET /api/admin/scenarios/{id}
	getScenario, _ := r.NewOperationContext(http.MethodGet, "/api/admin/scenarios/{id}")
	getScenario.SetSummary("Get scenario")
	getScenario.SetDescription("Returns a scenario with full stage details. Requires admin_session cookie.")
	getScenario.AddRespStructure(AdminScenarioDetail{}, openapi.WithHTTPStatus(http.StatusOK))
	getScenario.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusNotFound))
	getScenario.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(getScenario)

	// PUT /api/admin/scenarios/{id}
	updateScenario, _ := r.NewOperationContext(http.MethodPut, "/api/admin/scenarios/{id}")
	updateScenario.SetSummary("Update scenario")
	updateScenario.SetDescription("Updates a scenario and its stages. Requires admin_session cookie.")
	updateScenario.AddReqStructure(AdminScenarioRequest{})
	updateScenario.AddRespStructure(AdminScenarioDetail{}, openapi.WithHTTPStatus(http.StatusOK))
	updateScenario.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusBadRequest))
	updateScenario.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusNotFound))
	updateScenario.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(updateScenario)

	// DELETE /api/admin/scenarios/{id}
	deleteScenario, _ := r.NewOperationContext(http.MethodDelete, "/api/admin/scenarios/{id}")
	deleteScenario.SetSummary("Delete scenario")
	deleteScenario.SetDescription("Deletes a scenario. Blocked if games reference it. Requires admin_session cookie.")
	deleteScenario.AddRespStructure(nil, openapi.WithHTTPStatus(http.StatusOK))
	deleteScenario.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusConflict))
	deleteScenario.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusNotFound))
	deleteScenario.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(deleteScenario)

	// GET /api/admin/games
	listGames, _ := r.NewOperationContext(http.MethodGet, "/api/admin/games")
	listGames.SetSummary("List games")
	listGames.SetDescription("Returns all games with scenario names and team counts. Requires admin_session cookie.")
	listGames.AddRespStructure([]AdminGameSummary{}, openapi.WithHTTPStatus(http.StatusOK))
	listGames.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(listGames)

	// POST /api/admin/games
	createGame, _ := r.NewOperationContext(http.MethodPost, "/api/admin/games")
	createGame.SetSummary("Create game")
	createGame.SetDescription("Creates a new game for the demo client. Requires admin_session cookie.")
	createGame.AddReqStructure(AdminGameRequest{})
	createGame.AddRespStructure(AdminGameDetail{}, openapi.WithHTTPStatus(http.StatusCreated))
	createGame.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusBadRequest))
	createGame.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(createGame)

	// GET /api/admin/games/{gameID}
	getGame, _ := r.NewOperationContext(http.MethodGet, "/api/admin/games/{gameID}")
	getGame.SetSummary("Get game")
	getGame.SetDescription("Returns a game with teams and player counts. Requires admin_session cookie.")
	getGame.AddRespStructure(AdminGameDetail{}, openapi.WithHTTPStatus(http.StatusOK))
	getGame.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusNotFound))
	getGame.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(getGame)

	// PUT /api/admin/games/{gameID}
	updateGame, _ := r.NewOperationContext(http.MethodPut, "/api/admin/games/{gameID}")
	updateGame.SetSummary("Update game")
	updateGame.SetDescription("Updates a game's scenario, status, and timer. Requires admin_session cookie.")
	updateGame.AddReqStructure(AdminGameRequest{})
	updateGame.AddRespStructure(AdminGameDetail{}, openapi.WithHTTPStatus(http.StatusOK))
	updateGame.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusBadRequest))
	updateGame.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusNotFound))
	updateGame.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(updateGame)

	// DELETE /api/admin/games/{gameID}
	deleteGame, _ := r.NewOperationContext(http.MethodDelete, "/api/admin/games/{gameID}")
	deleteGame.SetSummary("Delete game")
	deleteGame.SetDescription("Deletes a game. Blocked if any team has players. Requires admin_session cookie.")
	deleteGame.AddRespStructure(nil, openapi.WithHTTPStatus(http.StatusOK))
	deleteGame.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusConflict))
	deleteGame.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusNotFound))
	deleteGame.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(deleteGame)

	// GET /api/admin/games/{gameID}/teams
	listTeams, _ := r.NewOperationContext(http.MethodGet, "/api/admin/games/{gameID}/teams")
	listTeams.SetSummary("List teams")
	listTeams.SetDescription("Returns teams for a game with player counts. Requires admin_session cookie.")
	listTeams.AddRespStructure([]AdminTeamItem{}, openapi.WithHTTPStatus(http.StatusOK))
	listTeams.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusNotFound))
	listTeams.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(listTeams)

	// POST /api/admin/games/{gameID}/teams
	createTeam, _ := r.NewOperationContext(http.MethodPost, "/api/admin/games/{gameID}/teams")
	createTeam.SetSummary("Create team")
	createTeam.SetDescription("Creates a team in a game. Auto-generates join token if blank. Requires admin_session cookie.")
	createTeam.AddReqStructure(AdminTeamRequest{})
	createTeam.AddRespStructure(AdminTeamItem{}, openapi.WithHTTPStatus(http.StatusCreated))
	createTeam.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusBadRequest))
	createTeam.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusConflict))
	createTeam.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(createTeam)

	// PUT /api/admin/games/{gameID}/teams/{teamID}
	updateTeam, _ := r.NewOperationContext(http.MethodPut, "/api/admin/games/{gameID}/teams/{teamID}")
	updateTeam.SetSummary("Update team")
	updateTeam.SetDescription("Updates a team's name and guide name. Token is immutable. Requires admin_session cookie.")
	updateTeam.AddReqStructure(AdminTeamRequest{})
	updateTeam.AddRespStructure(AdminTeamItem{}, openapi.WithHTTPStatus(http.StatusOK))
	updateTeam.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusBadRequest))
	updateTeam.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusNotFound))
	updateTeam.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(updateTeam)

	// DELETE /api/admin/games/{gameID}/teams/{teamID}
	deleteTeamOp, _ := r.NewOperationContext(http.MethodDelete, "/api/admin/games/{gameID}/teams/{teamID}")
	deleteTeamOp.SetSummary("Delete team")
	deleteTeamOp.SetDescription("Deletes a team. Blocked if players exist. Requires admin_session cookie.")
	deleteTeamOp.AddRespStructure(nil, openapi.WithHTTPStatus(http.StatusOK))
	deleteTeamOp.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusConflict))
	deleteTeamOp.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusNotFound))
	deleteTeamOp.AddRespStructure(ErrorResponse{}, openapi.WithHTTPStatus(http.StatusUnauthorized))
	_ = r.AddOperation(deleteTeamOp)

	return r.Spec
}

func handleOpenAPI() http.HandlerFunc {
	spec := newOpenAPISpec()
	data, _ := json.MarshalIndent(spec, "", "  ")

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}
