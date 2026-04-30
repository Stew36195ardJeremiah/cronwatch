package alert_test

import (
	"bytes"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/alert"
)

func TestLogNotifier_Send_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	log.SetFlags(0) // suppress date/time prefix for deterministic output
	t.Cleanup(func() { log.SetOutput(nil); log.SetFlags(log.LstdFlags) })

	notifier := &alert.LogNotifier{}
	a := alert.Alert{
		JobName:   "cleanup-job",
		Level:     alert.LevelError,
		Message:   "job overdue by 5m",
		Timestamp: time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
	}

	if err := notifier.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("expected output to contain ERROR, got: %s", output)
	}
	if !strings.Contains(output, "cleanup-job") {
		t.Errorf("expected output to contain job name, got: %s", output)
	}
	if !strings.Contains(output, "job overdue by 5m") {
		t.Errorf("expected output to contain message, got: %s", output)
	}
}
