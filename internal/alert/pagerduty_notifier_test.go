package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPagerDutyNotifier_Send_Success(t *testing.T) {
	var received pdPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := NewPagerDutyNotifier("test-key")
	n.eventURL = ts.URL

	a := Alert{Job: "backup", Message: "overdue by 5m", Level: LevelError, Time: time.Now()}
	if err := n.Send(a); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if received.RoutingKey != "test-key" {
		t.Errorf("expected routing key 'test-key', got %q", received.RoutingKey)
	}
	if received.Payload.Severity != "critical" {
		t.Errorf("expected severity 'critical', got %q", received.Payload.Severity)
	}
	if received.EventAction != "trigger" {
		t.Errorf("expected event_action 'trigger', got %q", received.EventAction)
	}
}

func TestPagerDutyNotifier_Send_WarnLevel(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p pdPayload
		_ = json.NewDecoder(r.Body).Decode(&p)
		if p.Payload.Severity != "warning" {
			t.Errorf("expected severity 'warning', got %q", p.Payload.Severity)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := NewPagerDutyNotifier("key")
	n.eventURL = ts.URL

	a := Alert{Job: "sync", Message: "drift detected", Level: LevelWarn, Time: time.Now()}
	if err := n.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPagerDutyNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	n := NewPagerDutyNotifier("key")
	n.eventURL = ts.URL

	a := Alert{Job: "job", Message: "fail", Level: LevelError, Time: time.Now()}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestPagerDutyNotifier_Send_BadURL(t *testing.T) {
	n := NewPagerDutyNotifier("key")
	n.eventURL = "http://127.0.0.1:0"

	a := Alert{Job: "job", Message: "fail", Level: LevelWarn, Time: time.Now()}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}
