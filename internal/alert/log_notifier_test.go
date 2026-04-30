package alert

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogNotifier_Send_WritesOutput(t *testing.T) {
	var buf bytes.Buffer
	notifier := NewLogNotifier(&buf)

	err := notifier.Send("WARN", "backup-job", "execution overdue by 5m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Errorf("expected output to contain level WARN, got: %s", output)
	}
	if !strings.Contains(output, "backup-job") {
		t.Errorf("expected output to contain job name, got: %s", output)
	}
	if !strings.Contains(output, "execution overdue by 5m") {
		t.Errorf("expected output to contain message, got: %s", output)
	}
}

func TestLogNotifier_Send_ErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	notifier := NewLogNotifier(&buf)

	err := notifier.Send("ERROR", "sync-job", "job failed to execute")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("expected output to contain level ERROR, got: %s", output)
	}
	if !strings.Contains(output, "sync-job") {
		t.Errorf("expected output to contain job name, got: %s", output)
	}
}

func TestLogNotifier_DefaultsToStdout(t *testing.T) {
	notifier := NewLogNotifier(nil)
	if notifier.Writer == nil {
		t.Error("expected Writer to default to os.Stdout, got nil")
	}
}
