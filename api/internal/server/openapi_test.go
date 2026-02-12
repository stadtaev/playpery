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
	if !strings.Contains(body, `"openapi": "3.1.0"`) {
		t.Fatalf("body missing openapi version")
	}
	if !strings.Contains(body, `"/healthz"`) {
		t.Fatalf("body missing /healthz path")
	}
}

func TestHandleSwaggerUI(t *testing.T) {
	h := handleSwaggerUI()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	rec := httptest.NewRecorder()

	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "text/html") {
		t.Fatalf("content-type = %q, want text/html", got)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "SwaggerUIBundle") {
		t.Fatalf("body missing SwaggerUIBundle")
	}
	if !strings.Contains(body, "/openapi.json") {
		t.Fatalf("body missing /openapi.json")
	}
}
