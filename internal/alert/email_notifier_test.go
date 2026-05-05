package alert

import (
	"fmt"
	"net/smtp"
	"strings"
	"testing"
)

func newTestEmailNotifier(dialFn func(string, smtp.Auth, string, []string, []byte) error) *EmailNotifier {
	n := NewEmailNotifier(EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "secret",
		From:     "alerts@example.com",
		To:       []string{"ops@example.com"},
	})
	n.dial = dialFn
	return n
}

func TestEmailNotifier_Send_Success(t *testing.T) {
	var capturedMsg []byte
	n := newTestEmailNotifier(func(_ string, _ smtp.Auth, _ string, _ []string, msg []byte) error {
		capturedMsg = msg
		return nil
	})

	err := n.Send(Alert{Job: "backup", Level: LevelError, Message: "job failed"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(string(capturedMsg), "job failed") {
		t.Errorf("expected message body to contain 'job failed'")
	}
	if !strings.Contains(string(capturedMsg), "[ERROR]") {
		t.Errorf("expected subject to contain '[ERROR]'")
	}
}

func TestEmailNotifier_Send_WarnLevel(t *testing.T) {
	var capturedMsg []byte
	n := newTestEmailNotifier(func(_ string, _ smtp.Auth, _ string, _ []string, msg []byte) error {
		capturedMsg = msg
		return nil
	})

	err := n.Send(Alert{Job: "sync", Level: LevelWarn, Message: "drift detected"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !strings.Contains(string(capturedMsg), "[WARN]") {
		t.Errorf("expected subject to contain '[WARN]'")
	}
}

func TestEmailNotifier_Send_DialError(t *testing.T) {
	n := newTestEmailNotifier(func(_ string, _ smtp.Auth, _ string, _ []string, _ []byte) error {
		return fmt.Errorf("connection refused")
	})

	err := n.Send(Alert{Job: "backup", Level: LevelError, Message: "failed"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("expected error to mention connection refused, got: %v", err)
	}
}

func TestEmailNotifier_Send_NoRecipients(t *testing.T) {
	n := NewEmailNotifier(EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		From:     "alerts@example.com",
		To:       []string{},
	})

	err := n.Send(Alert{Job: "backup", Level: LevelError, Message: "failed"})
	if err == nil {
		t.Fatal("expected error for missing recipients, got nil")
	}
}
