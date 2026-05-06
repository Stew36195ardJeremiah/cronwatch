package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func newTestRunLog() *monitor.RunLog {
	return monitor.NewRunLog(50)
}

func TestHandleRunLog_MissingJobParam(t *testing.T) {
	rl := newTestRunLog()
	h := newRunLogHandler(rl)

	req := httptest.NewRequest(http.MethodGet, "/api/runlog", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestHandleRunLog_WrongMethod(t *testing.T) {
	rl := newTestRunLog()
	h := newRunLogHandler(rl)

	req := httptest.NewRequest(http.MethodPost, "/api/runlog?job=backup", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}

func TestHandleRunLog_EmptyForUnknownJob(t *testing.T) {
	rl := newTestRunLog()
	h := newRunLogHandler(rl)

	req := httptest.NewRequest(http.MethodGet, "/api/runlog?job=unknown", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var entries []monitor.RunEntry
	if err := json.NewDecoder(rw.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty slice, got %d entries", len(entries))
	}
}

func TestHandleRunLog_ReturnsEntriesForJob(t *testing.T) {
	rl := newTestRunLog()
	h := newRunLogHandler(rl)

	now := time.Now()
	rl.Append(monitor.RunEntry{Job: "backup", StartedAt: now, Duration: 2 * time.Second, Success: true})
	rl.Append(monitor.RunEntry{Job: "backup", StartedAt: now.Add(-time.Minute), Duration: 3 * time.Second, Success: false})
	rl.Append(monitor.RunEntry{Job: "other", StartedAt: now, Duration: time.Second, Success: true})

	req := httptest.NewRequest(http.MethodGet, "/api/runlog?job=backup", nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var entries []monitor.RunEntry
	if err := json.NewDecoder(rw.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries for 'backup', got %d", len(entries))
	}
}

func TestHandleRunLog_LimitParam(t *testing.T) {
	rl := newTestRunLog()
	h := newRunLogHandler(rl)

	for i := 0; i < 10; i++ {
		rl.Append(monitor.RunEntry{Job: "sync", StartedAt: time.Now(), Duration: time.Second, Success: true})
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/runlog?job=sync&limit=%d", 3), nil)
	rw := httptest.NewRecorder()
	h.ServeHTTP(rw, req)

	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	var entries []monitor.RunEntry
	if err := json.NewDecoder(rw.Body).Decode(&entries); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries with limit=3, got %d", len(entries))
	}
}
