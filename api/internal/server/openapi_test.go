package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleOpenAPI(t *testing.T) {
	h := handleOpenAPI()
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	rec := httptest.NewRecorder()

	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("content-type = %q, want application/json", got)
	}

	body := rec.Body.String()
	if !strings.Contains(body, `"openapi": "3.0.3"`) {
		t.Fatalf("body missing openapi version:\n%s", body)
	}
	if !strings.Contains(body, `"/healthz"`) {
		t.Fatalf("body missing /healthz path")
	}
}
