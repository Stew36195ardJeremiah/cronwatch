package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cronwatch/internal/api"
	"github.com/cronwatch/internal/config"
	"github.com/cronwatch/internal/monitor"
)

func newTestMonitor(t *testing.T) *monitor.Monitor {
	t.Helper()
	cfg := &config.Config{
		Jobs: []config.Job{
			{Name: "test-job", Schedule: "@every 1m", Timeout: "30s"},
		},
	}
	m, err := monitor.New(cfg, nil)
	if err != nil {
		t.Fatalf("monitor.New: %v", err)
	}
	return m
}

func TestHandleHealth_ReturnsOK(t *testing.T) {
	srv := api.New(":0", newTestMonitor(t))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", body["status"])
	}
}

func TestHandleStatus_ReturnsArray(t *testing.T) {
	srv := api.New(":0", newTestMonitor(t))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/status", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

func TestHandleHistory_MissingJob(t *testing.T) {
	srv := api.New(":0", newTestMonitor(t))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleHistory_KnownJob(t *testing.T) {
	m := newTestMonitor(t)
	m.RecordRun("test-job", time.Now(), true)
	srv := api.New(":0", m)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/history?job=test-job", nil)
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body []map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body) == 0 {
		t.Error("expected at least one history record")
	}
}
