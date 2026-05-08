package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/monitor"
)

func newTestBaselineStore() *monitor.BaselineStore {
	s := monitor.NewBaselineStore(10)
	s.Record("job1", 2*time.Second)
	s.Record("job1", 4*time.Second)
	return s
}

func TestHandleBaselineAll_ReturnsEntries(t *testing.T) {
	h := newBaselineHandler(newTestBaselineStore())
	req := httptest.NewRequest(http.MethodGet, "/baseline", nil)
	w := httptest.NewRecorder()
	h.handleBaselineAll(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out []monitor.BaselineEntry
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(out))
	}
}

func TestHandleBaselineAll_WrongMethod(t *testing.T) {
	h := newBaselineHandler(newTestBaselineStore())
	req := httptest.NewRequest(http.MethodPost, "/baseline", nil)
	w := httptest.NewRecorder()
	h.handleBaselineAll(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestHandleBaselineGet_KnownJob(t *testing.T) {
	h := newBaselineHandler(newTestBaselineStore())
	req := httptest.NewRequest(http.MethodGet, "/baseline/job?job=job1", nil)
	w := httptest.NewRecorder()
	h.handleBaselineGet(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var e monitor.BaselineEntry
	if err := json.NewDecoder(w.Body).Decode(&e); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if e.JobName != "job1" {
		t.Fatalf("expected job1, got %s", e.JobName)
	}
}

func TestHandleBaselineGet_UnknownJob(t *testing.T) {
	h := newBaselineHandler(newTestBaselineStore())
	req := httptest.NewRequest(http.MethodGet, "/baseline/job?job=ghost", nil)
	w := httptest.NewRecorder()
	h.handleBaselineGet(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleBaselineGet_MissingParam(t *testing.T) {
	h := newBaselineHandler(newTestBaselineStore())
	req := httptest.NewRequest(http.MethodGet, "/baseline/job", nil)
	w := httptest.NewRecorder()
	h.handleBaselineGet(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleBaselineReset_Success(t *testing.T) {
	h := newBaselineHandler(newTestBaselineStore())
	req := httptest.NewRequest(http.MethodPost, "/baseline/reset?job=job1", nil)
	w := httptest.NewRecorder()
	h.handleBaselineReset(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
	_, ok := h.store.Get("job1")
	if ok {
		t.Fatal("expected baseline to be cleared")
	}
}
