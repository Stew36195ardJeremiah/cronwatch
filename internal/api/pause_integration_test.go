package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cronwatch/internal/monitor"
)

func TestPause_EndToEnd_PauseAndResume(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}

	store := monitor.NewPauseStore()
	h := newPauseHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/pause", h.handlePause)
	mux.HandleFunc("/resume", h.handleResume)
	mux.HandleFunc("/paused", h.handleListPaused)

	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go srv.ListenAndServe() //nolint:errcheck
	time.Sleep(50 * time.Millisecond)
	defer srv.Close()

	base := fmt.Sprintf("http://localhost:%d", port)

	// Pause a job
	body, _ := json.Marshal(map[string]string{"job": "nightly", "reason": "deploy", "paused_by": "ci"})
	resp, err := http.Post(base+"/pause", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("pause request: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	// List paused — should contain nightly
	resp, err = http.Get(base + "/paused")
	if err != nil {
		t.Fatalf("list request: %v", err)
	}
	var entries []map[string]any
	json.NewDecoder(resp.Body).Decode(&entries) //nolint:errcheck
	resp.Body.Close()
	if len(entries) != 1 {
		t.Fatalf("expected 1 paused job, got %d", len(entries))
	}

	// Resume
	body, _ = json.Marshal(map[string]string{"job": "nightly"})
	resp, err = http.Post(base+"/resume", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("resume request: %v", err)
	}
	resp.Body.Close()

	if store.IsPaused("nightly") {
		t.Fatal("expected nightly to be resumed")
	}
}
