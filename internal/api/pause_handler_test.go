package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cronwatch/internal/monitor"
)

func newTestPauseStore() *monitor.PauseStore {
	return monitor.NewPauseStore()
}

func postPause(t *testing.T, h *pauseHandler, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	rec := httptest.NewRecorder()
	switch path {
	case "/pause":
		h.handlePause(rec, req)
	case "/resume":
		h.handleResume(rec, req)
	}
	return rec
}

func TestHandlePause_Success(t *testing.T) {
	h := newPauseHandler(newTestPauseStore())
	rec := postPause(t, h, "/pause", map[string]string{"job": "backup", "reason": "maint", "paused_by": "alice"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !h.store.IsPaused("backup") {
		t.Fatal("expected job to be paused")
	}
}

func TestHandlePause_MissingJob(t *testing.T) {
	h := newPauseHandler(newTestPauseStore())
	rec := postPause(t, h, "/pause", map[string]string{"reason": "maint"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandlePause_WrongMethod(t *testing.T) {
	h := newPauseHandler(newTestPauseStore())
	req := httptest.NewRequest(http.MethodGet, "/pause", nil)
	rec := httptest.NewRecorder()
	h.handlePause(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleResume_Success(t *testing.T) {
	h := newPauseHandler(newTestPauseStore())
	h.store.Pause("sync", "", "")
	rec := postPause(t, h, "/resume", map[string]string{"job": "sync"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if h.store.IsPaused("sync") {
		t.Fatal("expected job to be resumed")
	}
}

func TestHandleListPaused_ReturnsList(t *testing.T) {
	h := newPauseHandler(newTestPauseStore())
	h.store.Pause("jobA", "", "")
	h.store.Pause("jobB", "", "")
	req := httptest.NewRequest(http.MethodGet, "/paused", nil)
	rec := httptest.NewRecorder()
	h.handleListPaused(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var result []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
}
