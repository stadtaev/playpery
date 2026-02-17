package server

import (
	"encoding/json"
	"net/http"

	openapi "github.com/swaggest/openapi-go"
	"github.com/swaggest/openapi-go/openapi3"
)

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

	// GET /ws/echo — WebSocket, description only
	getWSEcho, _ := r.NewOperationContext(http.MethodGet, "/ws/echo")
	getWSEcho.SetSummary("WebSocket echo")
	getWSEcho.SetDescription("Upgrades to a WebSocket connection that echoes messages back.")
	getWSEcho.AddRespStructure(nil, openapi.WithHTTPStatus(http.StatusSwitchingProtocols),
		openapi.WithContentType("text/plain"))
	_ = r.AddOperation(getWSEcho)

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
