package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func postSuppress(t *testing.T, s *Server, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/suppress", bytes.NewBufferString(body))
	s.handleSuppress(rec, req)
	return rec
}

func TestHandleSuppress_Success(t *testing.T) {
	s := &Server{monitor: newTestMonitor()}
	rec := postSuppress(t, s, `{"job":"backup","duration":"10m"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal("failed to decode response")
	}
	if resp["job"] != "backup" {
		t.Errorf("expected job=backup, got %v", resp["job"])
	}
}

func TestHandleSuppress_InvalidDuration(t *testing.T) {
	s := &Server{monitor: newTestMonitor()}
	rec := postSuppress(t, s, `{"job":"backup","duration":"not-a-duration"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleSuppress_MissingJob(t *testing.T) {
	s := &Server{monitor: newTestMonitor()}
	rec := postSuppress(t, s, `{"duration":"5m"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleSuppress_WrongMethod(t *testing.T) {
	s := &Server{monitor: newTestMonitor()}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/suppress", nil)
	s.handleSuppress(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleLiftSuppression_Success(t *testing.T) {
	s := &Server{monitor: newTestMonitor()}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/suppress/lift", bytes.NewBufferString(`{"job":"backup"}`))
	s.handleLiftSuppression(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
