package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSlackNotifier_Send_Success(t *testing.T) {
	var received slackPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	notifier := NewSlackNotifier(server.URL)
	err := notifier.Send(Alert{
		JobName: "backup",
		Level:   "ERROR",
		Message: "job overdue by 5m",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(received.Text, "backup") {
		t.Errorf("expected job name in message, got: %s", received.Text)
	}
	if !strings.Contains(received.Text, "ERROR") {
		t.Errorf("expected level in message, got: %s", received.Text)
	}
}

func TestSlackNotifier_Send_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	notifier := NewSlackNotifier(server.URL)
	err := notifier.Send(Alert{
		JobName: "cleanup",
		Level:   "WARN",
		Message: "drift detected",
	})

	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}

func TestSlackNotifier_Send_BadURL(t *testing.T) {
	notifier := NewSlackNotifier("http://127.0.0.1:0/invalid")
	err := notifier.Send(Alert{
		JobName: "test",
		Level:   "ERROR",
		Message: "unreachable",
	})

	if err == nil {
		t.Fatal("expected error for bad URL, got nil")
	}
}
