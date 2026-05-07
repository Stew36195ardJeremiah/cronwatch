package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/monitor"
)

func newTestSilenceStore() *monitor.SilenceWindowStore {
	return monitor.NewSilenceWindowStore()
}

func postSilence(t *testing.T, h *silenceWindowHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/silence", bytes.NewBufferString(body))
	h.handleAdd(rec, req)
	return rec
}

func TestHandleSilenceAdd_Success(t *testing.T) {
	h := newSilenceWindowHandler(newTestSilenceStore())
	now := time.Now().UTC()
	body, _ := json.Marshal(map[string]string{
		"job":    "backup",
		"start":  now.Format(time.RFC3339),
		"end":    now.Add(time.Hour).Format(time.RFC3339),
		"reason": "planned maintenance",
	})
	rec := postSilence(t, h, string(body))
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
}

func TestHandleSilenceAdd_MissingFields(t *testing.T) {
	h := newSilenceWindowHandler(newTestSilenceStore())
	rec := postSilence(t, h, `{"job":"backup"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleSilenceAdd_InvalidJSON(t *testing.T) {
	h := newSilenceWindowHandler(newTestSilenceStore())
	rec := postSilence(t, h, `not-json`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleSilenceAdd_WrongMethod(t *testing.T) {
	h := newSilenceWindowHandler(newTestSilenceStore())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/silence", nil)
	h.handleAdd(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleSilenceCheck_Silenced(t *testing.T) {
	store := newTestSilenceStore()
	now := time.Now()
	store.Add(monitor.SilenceWindow{
		JobName: "nightly",
		Start:   now.Add(-time.Minute),
		End:     now.Add(time.Hour),
	})
	h := newSilenceWindowHandler(store)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/silence/check?job=nightly", nil)
	h.handleCheck(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]bool
	json.NewDecoder(rec.Body).Decode(&resp)
	if !resp["silenced"] {
		t.Fatal("expected silenced=true")
	}
}

func TestHandleSilenceList_ReturnsAll(t *testing.T) {
	store := newTestSilenceStore()
	now := time.Now()
	store.Add(monitor.SilenceWindow{JobName: "job-x", Start: now, End: now.Add(time.Hour)})
	h := newSilenceWindowHandler(store)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/silence", nil)
	h.handleList(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]interface{}
	json.NewDecoder(rec.Body).Decode(&resp)
	if _, ok := resp["job-x"]; !ok {
		t.Fatal("expected job-x in response")
	}
}
