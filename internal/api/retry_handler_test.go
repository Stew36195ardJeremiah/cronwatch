package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func newTestRetryStore() *monitor.RetryStore {
	return monitor.NewRetryStore(monitor.RetryPolicy{MaxAttempts: 3, Interval: 0})
}

func TestHandleRetryStatus_NotFound(t *testing.T) {
	store := newTestRetryStore()
	h := handleRetryStatus(store)
	req := httptest.NewRequest(http.MethodGet, "/retry/status?job=ghost", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleRetryStatus_MissingParam(t *testing.T) {
	store := newTestRetryStore()
	h := handleRetryStatus(store)
	req := httptest.NewRequest(http.MethodGet, "/retry/status", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleRetryStatus_ReturnsState(t *testing.T) {
	store := monitor.NewRetryStore(monitor.RetryPolicy{MaxAttempts: 5, Interval: time.Millisecond})
	store.ShouldRetry("backup")
	h := handleRetryStatus(store)
	req := httptest.NewRequest(http.MethodGet, "/retry/status?job=backup", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	body := rr.Body.String()
	if body == "" {
		t.Error("expected non-empty body")
	}
}

func TestHandleRetryStatus_WrongMethod(t *testing.T) {
	store := newTestRetryStore()
	h := handleRetryStatus(store)
	req := httptest.NewRequest(http.MethodPost, "/retry/status?job=x", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHandleRetryReset_Success(t *testing.T) {
	store := newTestRetryStore()
	store.ShouldRetry("cleanup")
	h := handleRetryReset(store)
	req := httptest.NewRequest(http.MethodPost, "/retry/reset?job=cleanup", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
	_, ok := store.State("cleanup")
	if ok {
		t.Error("expected state to be cleared after reset")
	}
}

func TestHandleRetryReset_MissingJob(t *testing.T) {
	store := newTestRetryStore()
	h := handleRetryReset(store)
	req := httptest.NewRequest(http.MethodPost, "/retry/reset", nil)
	rr := httptest.NewRecorder()
	h(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
