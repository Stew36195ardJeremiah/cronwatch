package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwatch/cronwatch/internal/monitor"
)

func newTestRoutingStore() *monitor.RoutingStore {
	return monitor.NewRoutingStore("log")
}

func TestHandleSetRoute_Success(t *testing.T) {
	h := newRoutingHandler(newTestRoutingStore())
	body := `{"job":"backup","channel":"slack"}`
	req := httptest.NewRequest(http.MethodPost, "/routes", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.handleSetRoute(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestHandleSetRoute_MissingFields(t *testing.T) {
	h := newRoutingHandler(newTestRoutingStore())
	body := `{"job":"","channel":"slack"}`
	req := httptest.NewRequest(http.MethodPost, "/routes", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	h.handleSetRoute(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleSetRoute_WrongMethod(t *testing.T) {
	h := newRoutingHandler(newTestRoutingStore())
	req := httptest.NewRequest(http.MethodGet, "/routes", nil)
	rec := httptest.NewRecorder()
	h.handleSetRoute(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleDeleteRoute_Success(t *testing.T) {
	store := newTestRoutingStore()
	_ = store.Set("myjob", "email")
	h := newRoutingHandler(store)
	req := httptest.NewRequest(http.MethodDelete, "/routes?job=myjob", nil)
	rec := httptest.NewRecorder()
	h.handleDeleteRoute(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if store.Resolve("myjob") != "log" {
		t.Fatal("expected job to fall back to default after delete")
	}
}

func TestHandleDeleteRoute_MissingParam(t *testing.T) {
	h := newRoutingHandler(newTestRoutingStore())
	req := httptest.NewRequest(http.MethodDelete, "/routes", nil)
	rec := httptest.NewRecorder()
	h.handleDeleteRoute(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleListRoutes_ReturnsRoutes(t *testing.T) {
	store := newTestRoutingStore()
	_ = store.Set("job-a", "slack")
	h := newRoutingHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/routes", nil)
	rec := httptest.NewRecorder()
	h.handleListRoutes(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp struct {
		Default string `json:"default"`
		Routes  []struct {
			Job     string `json:"Job"`
			Channel string `json:"Channel"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Default != "log" {
		t.Fatalf("expected default 'log', got %q", resp.Default)
	}
	if len(resp.Routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(resp.Routes))
	}
}
