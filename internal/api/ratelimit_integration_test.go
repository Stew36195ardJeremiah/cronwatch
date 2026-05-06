package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func TestRateLimit_EndToEnd_StatusAndReset(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}

	store := monitor.NewRateLimitStore(time.Minute, 3)
	store.Allow("integration-job")
	store.Allow("integration-job")

	cfg := buildTestConfig()
	srv := New(cfg, buildTestMonitor(cfg))
	rlh := newRateLimitHandler(store)
	srv.mux.HandleFunc("/api/ratelimit/status", rlh.handleRateLimitStatus)
	srv.mux.HandleFunc("/api/ratelimit/reset", rlh.handleRateLimitReset)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	go srv.Start(addr) //nolint:errcheck
	time.Sleep(50 * time.Millisecond)

	// Check status
	resp, err := http.Get(fmt.Sprintf("http://%s/api/ratelimit/status", addr))
	if err != nil {
		t.Fatalf("GET status: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var entries []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	// Reset
	resetResp, err := http.Post(
		fmt.Sprintf("http://%s/api/ratelimit/reset?job=integration-job", addr),
		"application/json", nil,
	)
	if err != nil {
		t.Fatalf("POST reset: %v", err)
	}
	defer resetResp.Body.Close()
	if resetResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 on reset, got %d", resetResp.StatusCode)
	}

	// Confirm cleared
	if _, ok := store.All()["integration-job"]; ok {
		t.Error("expected rate limit entry to be cleared")
	}

	srv.Stop()
}
