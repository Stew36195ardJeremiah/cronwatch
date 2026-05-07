package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/monitor"
)

func TestRouting_EndToEnd_SetAndResolve(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatalf("could not get free port: %v", err)
	}

	store := monitor.NewRoutingStore("log")
	rh := newRoutingHandler(store)

	m := newTestMonitor()
	srv := New(m, fmt.Sprintf(":%d", port))
	srv.mux.HandleFunc("/routes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			rh.handleSetRoute(w, r)
		case http.MethodGet:
			rh.handleListRoutes(w, r)
		case http.MethodDelete:
			rh.handleDeleteRoute(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	go srv.Start()
	defer srv.Stop()
	time.Sleep(50 * time.Millisecond)

	base := fmt.Sprintf("http://localhost:%d", port)

	// Set a route
	body := `{"job":"nightly-backup","channel":"pagerduty"}`
	resp, err := http.Post(base+"/routes", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("POST /routes failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp.StatusCode)
	}

	// List routes
	resp, err = http.Get(base + "/routes")
	if err != nil {
		t.Fatalf("GET /routes failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var result struct {
		Default string `json:"default"`
		Routes  []struct {
			Job     string `json:"Job"`
			Channel string `json:"Channel"`
		} `json:"routes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(result.Routes) != 1 || result.Routes[0].Job != "nightly-backup" {
		t.Fatalf("unexpected routes: %+v", result.Routes)
	}
}
