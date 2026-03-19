package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestParsePositiveInt(t *testing.T) {
	if got := parsePositiveInt("", 10); got != 10 {
		t.Fatalf("empty should fallback: got %d", got)
	}
	if got := parsePositiveInt("abc", 10); got != 10 {
		t.Fatalf("invalid should fallback: got %d", got)
	}
	if got := parsePositiveInt("-1", 10); got != 10 {
		t.Fatalf("non-positive should fallback: got %d", got)
	}
	if got := parsePositiveInt("25", 10); got != 25 {
		t.Fatalf("valid should parse: got %d", got)
	}
}

func TestParseIDInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drives/not-a-number", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "not-a-number")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	_, ok := parseID(rec, req)
	if ok {
		t.Fatal("expected parseID to fail for invalid id")
	}
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"error":"invalid id"`) {
		t.Fatalf("expected invalid id JSON error, got %q", rec.Body.String())
	}
}

func TestEventsHandlerUnavailable(t *testing.T) {
	h := &Handlers{events: nil}
	req := httptest.NewRequest(http.MethodGet, "/api/v1/events", nil)
	rec := httptest.NewRecorder()

	h.Events(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "event stream unavailable") {
		t.Fatalf("expected event stream unavailable message, got %q", rec.Body.String())
	}
}

func TestHealthz(t *testing.T) {
	h := &Handlers{}
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()

	h.Healthz(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected ok status, got %q", rec.Body.String())
	}
}

func TestReadyzStorageUnavailable(t *testing.T) {
	h := &Handlers{}
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rec := httptest.NewRecorder()

	h.Readyz(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"error":"storage unavailable"`) {
		t.Fatalf("expected storage unavailable message, got %q", rec.Body.String())
	}
}
