package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/monitor"
)

func newTestEscalationStore() *monitor.EscalationStore {
	return monitor.NewEscalationStore()
}

func TestHandleSetPolicy_Success(t *testing.T) {
	h := newEscalationHandler(newTestEscalationStore())
	body := `{"job":"backup","warn_after":"5m","critical_after":"15m"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/escalation/policy", bytes.NewBufferString(body))
	h.handleSetPolicy(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestHandleSetPolicy_MissingJob(t *testing.T) {
	h := newEscalationHandler(newTestEscalationStore())
	body := `{"warn_after":"5m"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/escalation/policy", bytes.NewBufferString(body))
	h.handleSetPolicy(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleSetPolicy_InvalidDuration(t *testing.T) {
	h := newEscalationHandler(newTestEscalationStore())
	body := `{"job":"sync","warn_after":"notaduration"}`
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/escalation/policy", bytes.NewBufferString(body))
	h.handleSetPolicy(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleSetPolicy_WrongMethod(t *testing.T) {
	h := newEscalationHandler(newTestEscalationStore())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/escalation/policy", nil)
	h.handleSetPolicy(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleStatus_ReturnsLevels(t *testing.T) {
	es := newTestEscalationStore()
	es.SetPolicy(monitor.EscalationPolicy{
		JobName:   "nightly",
		WarnAfter: time.Millisecond,
	})
	es.Trigger("nightly")
	time.Sleep(5 * time.Millisecond)

	h := newEscalationHandler(es)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/escalation/status", nil)
	h.handleStatus(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var out map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&out); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if out["nightly"] != "warn" {
		t.Fatalf("expected warn for nightly, got %s", out["nightly"])
	}
}

func TestHandleReset_ClearsEscalation(t *testing.T) {
	es := newTestEscalationStore()
	es.SetPolicy(monitor.EscalationPolicy{
		JobName:   "cleanup",
		WarnAfter: time.Millisecond,
	})
	es.Trigger("cleanup")

	h := newEscalationHandler(es)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/escalation/reset?job=cleanup", nil)
	h.handleReset(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if lvl := es.Level("cleanup"); lvl != monitor.EscalationNone {
		t.Fatalf("expected None after reset, got %s", lvl)
	}
}

func TestHandleReset_MissingJobParam(t *testing.T) {
	h := newEscalationHandler(newTestEscalationStore())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/escalation/reset", nil)
	h.handleReset(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}
