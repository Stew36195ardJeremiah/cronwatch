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

func newTestQuotaStore() *monitor.QuotaStore {
	return monitor.NewQuotaStore(5, time.Minute)
}

func TestHandleQuotaStatus_WrongMethod(t *testing.T) {
	h := newQuotaHandler(newTestQuotaStore())
	rec := httptest.NewRecorder()
	h.handleQuotaStatus(rec, httptest.NewRequest(http.MethodPost, "/quota/status?job=x", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rec.Code)
	}
}

func TestHandleQuotaStatus_MissingJob(t *testing.T) {
	h := newQuotaHandler(newTestQuotaStore())
	rec := httptest.NewRecorder()
	h.handleQuotaStatus(rec, httptest.NewRequest(http.MethodGet, "/quota/status", nil))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleQuotaStatus_NotFound(t *testing.T) {
	h := newQuotaHandler(newTestQuotaStore())
	rec := httptest.NewRecorder()
	h.handleQuotaStatus(rec, httptest.NewRequest(http.MethodGet, "/quota/status?job=unknown", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestHandleQuotaStatus_ReturnsEntry(t *testing.T) {
	store := newTestQuotaStore()
	store.Allow("backup")
	h := newQuotaHandler(store)
	rec := httptest.NewRecorder()
	h.handleQuotaStatus(rec, httptest.NewRequest(http.MethodGet, "/quota/status?job=backup", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var entry monitor.QuotaEntry
	if err := json.NewDecoder(rec.Body).Decode(&entry); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if entry.Count != 1 {
		t.Errorf("expected count 1, got %d", entry.Count)
	}
}

func TestHandleQuotaSet_Success(t *testing.T) {
	h := newQuotaHandler(newTestQuotaStore())
	body, _ := json.Marshal(map[string]interface{}{"job": "deploy", "limit": 3, "window": "5m"})
	rec := httptest.NewRecorder()
	h.handleQuotaSet(rec, httptest.NewRequest(http.MethodPost, "/quota/set", bytes.NewReader(body)))
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
}

func TestHandleQuotaSet_InvalidDuration(t *testing.T) {
	h := newQuotaHandler(newTestQuotaStore())
	body, _ := json.Marshal(map[string]interface{}{"job": "deploy", "limit": 3, "window": "bad"})
	rec := httptest.NewRecorder()
	h.handleQuotaSet(rec, httptest.NewRequest(http.MethodPost, "/quota/set", bytes.NewReader(body)))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestHandleQuotaReset_Success(t *testing.T) {
	store := newTestQuotaStore()
	store.Allow("cleanup")
	h := newQuotaHandler(store)
	rec := httptest.NewRecorder()
	h.handleQuotaReset(rec, httptest.NewRequest(http.MethodPost, "/quota/reset?job=cleanup", nil))
	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}
	e := store.Get("cleanup")
	if e == nil || e.Count != 0 {
		t.Error("expected count reset to 0")
	}
}
