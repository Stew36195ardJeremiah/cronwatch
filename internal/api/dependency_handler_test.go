package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func newTestDepStore() *monitor.DependencyStore {
	return monitor.NewDependencyStore()
}

func TestHandleAddEdge_Success(t *testing.T) {
	h := newDependencyHandler(newTestDepStore())
	body := `{"from":"job-a","to":"job-b"}`
	req := httptest.NewRequest(http.MethodPost, "/api/dependencies", bytes.NewBufferString(body))
	rw := httptest.NewRecorder()
	h.handleAddEdge(rw, req)
	if rw.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rw.Code)
	}
}

func TestHandleAddEdge_MissingFields(t *testing.T) {
	h := newDependencyHandler(newTestDepStore())
	body := `{"from":"job-a"}`
	req := httptest.NewRequest(http.MethodPost, "/api/dependencies", bytes.NewBufferString(body))
	rw := httptest.NewRecorder()
	h.handleAddEdge(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestHandleAddEdge_WrongMethod(t *testing.T) {
	h := newDependencyHandler(newTestDepStore())
	req := httptest.NewRequest(http.MethodGet, "/api/dependencies", nil)
	rw := httptest.NewRecorder()
	h.handleAddEdge(rw, req)
	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}

func TestHandleListEdges_ReturnsEdges(t *testing.T) {
	ds := newTestDepStore()
	ds.AddEdge("a", "b")
	h := newDependencyHandler(ds)
	req := httptest.NewRequest(http.MethodGet, "/api/dependencies", nil)
	rw := httptest.NewRecorder()
	h.handleListEdges(rw, req)
	var edges []monitor.DependencyEdge
	if err := json.NewDecoder(rw.Body).Decode(&edges); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}
}

func TestHandleBlocked_NotBlocked(t *testing.T) {
	ds := newTestDepStore()
	ds.AddEdge("job-a", "job-b")
	ds.MarkSuccess("job-a", time.Now())
	h := newDependencyHandler(ds)
	req := httptest.NewRequest(http.MethodGet, "/api/dependencies/blocked?job=job-b&since=1h", nil)
	rw := httptest.NewRecorder()
	h.handleBlocked(rw, req)
	var result map[string]interface{}
	json.NewDecoder(rw.Body).Decode(&result)
	if result["blocked"].(bool) {
		t.Fatal("expected job-b to not be blocked")
	}
}

func TestHandleBlocked_MissingJobParam(t *testing.T) {
	h := newDependencyHandler(newTestDepStore())
	req := httptest.NewRequest(http.MethodGet, "/api/dependencies/blocked", nil)
	rw := httptest.NewRecorder()
	h.handleBlocked(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}
