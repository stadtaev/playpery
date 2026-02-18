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
