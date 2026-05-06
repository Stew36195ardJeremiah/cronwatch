package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func newTestRateLimitStore() *monitor.RateLimitStore {
	return monitor.NewRateLimitStore(time.Minute, 3)
}

func TestHandleRateLimitStatus_WrongMethod(t *testing.T) {
	h := newRateLimitHandler(newTestRateLimitStore())
	req := httptest.NewRequest(http.MethodPost, "/ratelimit/status", nil)
	w := httptest.NewRecorder()
	h.handleRateLimitStatus(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleRateLimitStatus_EmptyStore(t *testing.T) {
	h := newRateLimitHandler(newTestRateLimitStore())
	req := httptest.NewRequest(http.MethodGet, "/ratelimit/status", nil)
	w := httptest.NewRecorder()
	h.handleRateLimitStatus(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var result []interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty array, got %d items", len(result))
	}
}

func TestHandleRateLimitStatus_ReturnsEntries(t *testing.T) {
	store := newTestRateLimitStore()
	store.Allow("backup-job")
	h := newRateLimitHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/ratelimit/status", nil)
	w := httptest.NewRecorder()
	h.handleRateLimitStatus(w, req)
	var result []map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(result))
	}
	if result[0]["job"] != "backup-job" {
		t.Errorf("unexpected job name: %v", result[0]["job"])
	}
}

func TestHandleRateLimitReset_WrongMethod(t *testing.T) {
	h := newRateLimitHandler(newTestRateLimitStore())
	req := httptest.NewRequest(http.MethodGet, "/ratelimit/reset?job=x", nil)
	w := httptest.NewRecorder()
	h.handleRateLimitReset(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHandleRateLimitReset_MissingJob(t *testing.T) {
	h := newRateLimitHandler(newTestRateLimitStore())
	req := httptest.NewRequest(http.MethodPost, "/ratelimit/reset", nil)
	w := httptest.NewRecorder()
	h.handleRateLimitReset(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleRateLimitReset_Success(t *testing.T) {
	store := newTestRateLimitStore()
	store.Allow("nightly-sync")
	h := newRateLimitHandler(store)
	req := httptest.NewRequest(http.MethodPost, "/ratelimit/reset?job=nightly-sync", nil)
	w := httptest.NewRecorder()
	h.handleRateLimitReset(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if _, ok := store.All()["nightly-sync"]; ok {
		t.Error("expected entry to be cleared after reset")
	}
}
