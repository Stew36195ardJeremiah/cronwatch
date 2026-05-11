package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/example/cronwatch/internal/monitor"
)

func newTestAuditStore() *monitor.AuditLogStore {
	s := monitor.NewAuditLogStore()
	s.Record("job1", monitor.AuditActionPause, "admin", "maintenance window")
	s.Record("job2", monitor.AuditActionSuppress, "ci", "deploy")
	s.Record("job1", monitor.AuditActionResume, "admin", "")
	return s
}

func TestHandleAuditAll_ReturnsEntries(t *testing.T) {
	store := newTestAuditStore()
	h := newAuditLogHandler(store)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit", nil)
	h.handleAuditAll(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var entries []monitor.AuditEntry
	if err := json.NewDecoder(rr.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("expected 3 entries, got %d", len(entries))
	}
}

func TestHandleAuditAll_WrongMethod(t *testing.T) {
	h := newAuditLogHandler(monitor.NewAuditLogStore())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/audit", nil)
	h.handleAuditAll(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHandleAuditForJob_ReturnsFiltered(t *testing.T) {
	store := newTestAuditStore()
	h := newAuditLogHandler(store)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/job?job=job1", nil)
	h.handleAuditForJob(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var entries []monitor.AuditEntry
	if err := json.NewDecoder(rr.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for job1, got %d", len(entries))
	}
	for _, e := range entries {
		if e.Job != "job1" {
			t.Errorf("unexpected job %q in filtered result", e.Job)
		}
	}
}

func TestHandleAuditForJob_MissingParam(t *testing.T) {
	h := newAuditLogHandler(monitor.NewAuditLogStore())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/audit/job", nil)
	h.handleAuditForJob(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleAuditForJob_WrongMethod(t *testing.T) {
	h := newAuditLogHandler(monitor.NewAuditLogStore())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/audit/job?job=x", nil)
	h.handleAuditForJob(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
