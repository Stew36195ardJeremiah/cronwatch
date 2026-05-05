package monitor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookNotifier_Send_Success(t *testing.T) {
	var received WebhookPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := NewWebhookNotifier(ts.URL)
	if err := n.Notify("backup", "overdue", "job is late"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if received.Job != "backup" {
		t.Errorf("expected job=backup, got %q", received.Job)
	}
	if received.Status != "overdue" {
		t.Errorf("expected status=overdue, got %q", received.Status)
	}
	if received.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestWebhookNotifier_Send_NonOKStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := NewWebhookNotifier(ts.URL)
	err := n.Notify("sync", "failed", "exit code 1")
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestWebhookNotifier_Send_BadURL(t *testing.T) {
	n := NewWebhookNotifier("http://127.0.0.1:0/no-listener")
	err := n.Notify("job", "error", "msg")
	if err == nil {
		t.Fatal("expected error for unreachable URL")
	}
}
