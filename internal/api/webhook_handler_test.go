package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func postIngest(t *testing.T, srv *Server, body any) *httptest.ResponseRecorder {
	t.Helper()
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()
	srv.handleIngestRun(rw, req)
	return rw
}

func TestHandleIngestRun_Success(t *testing.T) {
	m := newTestMonitor(t)
	srv := New(m)
	rw := postIngest(t, srv, RunEvent{Job: "daily-backup", Status: "success", Timestamp: time.Now()})
	if rw.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", rw.Code)
	}
}

func TestHandleIngestRun_MissingJob(t *testing.T) {
	m := newTestMonitor(t)
	srv := New(m)
	rw := postIngest(t, srv, RunEvent{Status: "success"})
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestHandleIngestRun_InvalidJSON(t *testing.T) {
	m := newTestMonitor(t)
	srv := New(m)
	req := httptest.NewRequest(http.MethodPost, "/ingest", bytes.NewBufferString("not-json"))
	rw := httptest.NewRecorder()
	srv.handleIngestRun(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rw.Code)
	}
}

func TestHandleIngestRun_WrongMethod(t *testing.T) {
	m := newTestMonitor(t)
	srv := New(m)
	req := httptest.NewRequest(http.MethodGet, "/ingest", nil)
	rw := httptest.NewRecorder()
	srv.handleIngestRun(rw, req)
	if rw.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rw.Code)
	}
}
