package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

func TestDependency_EndToEnd_AddAndCheck(t *testing.T) {
	port, err := freePort()
	if err != nil {
		t.Fatal(err)
	}

	cfg, mon := newTestMonitor()
	_ = cfg
	ds := monitor.NewDependencyStore()

	srv := New(mon)
	dh := newDependencyHandler(ds)
	srv.mux.HandleFunc("/api/dependencies", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			dh.handleAddEdge(w, r)
		} else {
			dh.handleListEdges(w, r)
		}
	})
	srv.mux.HandleFunc("/api/dependencies/blocked", dh.handleBlocked)

	addr := fmt.Sprintf("127.0.0.1:%d", port)
	go srv.Start(addr) //nolint:errcheck
	time.Sleep(50 * time.Millisecond)

	// Add edge
	body := `{"from":"ingest","to":"report"}`
	resp, err := http.Post("http://"+addr+"/api/dependencies", "application/json", bytes.NewBufferString(body))
	if err != nil {
		t.Fatalf("post edge: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", resp.StatusCode)
	}

	// Check blocked (upstream never ran)
	resp, err = http.Get("http://" + addr + "/api/dependencies/blocked?job=report&since=1h")
	if err != nil {
		t.Fatalf("get blocked: %v", err)
	}
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	if !result["blocked"].(bool) {
		t.Fatal("expected report to be blocked before ingest runs")
	}

	// Mark upstream success and re-check
	ds.MarkSuccess("ingest", time.Now())
	resp, err = http.Get("http://" + addr + "/api/dependencies/blocked?job=report&since=1h")
	if err != nil {
		t.Fatalf("get blocked: %v", err)
	}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	if result["blocked"].(bool) {
		t.Fatal("expected report to be unblocked after ingest succeeds")
	}

	srv.Stop()
}
