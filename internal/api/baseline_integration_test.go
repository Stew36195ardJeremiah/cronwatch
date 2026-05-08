package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/example/cronwatch/internal/monitor"
)

func TestBaseline_EndToEnd_RecordAndQuery(t *testing.T) {
	store := monitor.NewBaselineStore(10)
	store.Record("nightly-backup", 30*time.Second)
	store.Record("nightly-backup", 50*time.Second)

	h := newBaselineHandler(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/baseline", h.handleBaselineAll)
	mux.HandleFunc("/baseline/job", h.handleBaselineGet)
	mux.HandleFunc("/baseline/reset", h.handleBaselineReset)

	port, err := freePort()
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	srv := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go srv.ListenAndServe() //nolint:errcheck
	time.Sleep(30 * time.Millisecond)
	defer srv.Close()

	base := fmt.Sprintf("http://localhost:%d", port)

	// Query all
	resp, err := http.Get(base + "/baseline")
	if err != nil {
		t.Fatalf("GET /baseline: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var all []monitor.BaselineEntry
	if err := json.NewDecoder(resp.Body).Decode(&all); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(all))
	}
	if all[0].AvgDuration != 40*time.Second {
		t.Fatalf("expected avg=40s, got %s", all[0].AvgDuration)
	}

	// Reset and verify gone
	resp2, err := http.Post(base+"/baseline/reset?job=nightly-backup", "", nil)
	if err != nil {
		t.Fatalf("POST reset: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", resp2.StatusCode)
	}

	resp3, err := http.Get(base + "/baseline/job?job=nightly-backup")
	if err != nil {
		t.Fatalf("GET job: %v", err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 after reset, got %d", resp3.StatusCode)
	}
}
