package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/internal/monitor"
)

func newTestFlapStore() *monitor.FlapStore {
	return monitor.NewFlapStore(10*time.Minute, 3)
}

func TestHandleFlapStatus_MissingJob(t *testing.T) {
	h := newFlapHandler(newTestFlapStore())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/flap/status", nil)
	h.handleFlapStatus(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleFlapStatus_WrongMethod(t *testing.T) {
	h := newFlapHandler(newTestFlapStore())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/flap/status?job=j1", nil)
	h.handleFlapStatus(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHandleFlapStatus_UnknownJob(t *testing.T) {
	h := newFlapHandler(newTestFlapStore())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/flap/status?job=unknown", nil)
	h.handleFlapStatus(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&body)
	if body["is_flapping"].(bool) {
		t.Error("expected is_flapping=false for unknown job")
	}
}

func TestHandleFlapStatus_KnownFlappingJob(t *testing.T) {
	store := newTestFlapStore()
	now := time.Now()
	store.RecordTransition("job1", now.Add(-2*time.Minute))
	store.RecordTransition("job1", now.Add(-1*time.Minute))
	store.RecordTransition("job1", now)

	h := newFlapHandler(store)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/flap/status?job=job1", nil)
	h.handleFlapStatus(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var body map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&body)
	if !body["is_flapping"].(bool) {
		t.Error("expected is_flapping=true")
	}
}

func TestHandleFlapReset_Success(t *testing.T) {
	store := newTestFlapStore()
	now := time.Now()
	store.RecordTransition("job1", now.Add(-2*time.Minute))
	store.RecordTransition("job1", now.Add(-1*time.Minute))
	store.RecordTransition("job1", now)

	h := newFlapHandler(store)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/flap/reset?job=job1", nil)
	h.handleFlapReset(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if store.IsFlapping("job1") {
		t.Error("expected job1 not flapping after reset")
	}
}

func TestHandleFlapAll_ReturnsEntries(t *testing.T) {
	store := newTestFlapStore()
	now := time.Now()
	store.RecordTransition("jobA", now.Add(-2*time.Minute))
	store.RecordTransition("jobA", now.Add(-1*time.Minute))
	store.RecordTransition("jobA", now)

	h := newFlapHandler(store)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/flap/all", nil)
	h.handleFlapAll(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var rows []map[string]interface{}
	json.NewDecoder(rr.Body).Decode(&rows)
	if len(rows) != 1 {
		t.Errorf("expected 1 row, got %d", len(rows))
	}
}
