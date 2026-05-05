package api_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/cronwatch/internal/api"
)

func freePort(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	addr := l.Addr().String()
	l.Close()
	return addr
}

func TestServer_StartStop(t *testing.T) {
	addr := freePort(t)
	m := newTestMonitor(t)
	srv := api.New(addr, m)

	go func() { _ = srv.Start() }()
	time.Sleep(50 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://%s/healthz", addr))
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	if err := srv.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestServer_StatusEndpointLive(t *testing.T) {
	addr := freePort(t)
	m := newTestMonitor(t)
	srv := api.New(addr, m)

	go func() { _ = srv.Start() }()
	time.Sleep(50 * time.Millisecond)
	defer srv.Stop()

	resp, err := http.Get(fmt.Sprintf("http://%s/status", addr))
	if err != nil {
		t.Fatalf("GET /status: %v", err)
	}
	defer resp.Body.Close()

	var body []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
}
