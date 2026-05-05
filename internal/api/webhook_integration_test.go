package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestIngest_EndToEnd_RecordsRun(t *testing.T) {
	m := newTestMonitor(t)
	srv := New(m)

	port, err := freePort()
	if err != nil {
		t.Fatalf("freePort: %v", err)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	if err := srv.Start(addr); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer srv.Stop()

	time.Sleep(50 * time.Millisecond)

	event := RunEvent{Job: "weekly-report", Status: "success", Timestamp: time.Now()}
	body, _ := json.Marshal(event)
	resp, err := http.Post(
		fmt.Sprintf("http://%s/ingest", addr),
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatalf("POST /ingest: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}

	statuses := m.Statuses()
	if _, ok := statuses["weekly-report"]; !ok {
		t.Error("expected weekly-report to appear in statuses after ingest")
	}
}
