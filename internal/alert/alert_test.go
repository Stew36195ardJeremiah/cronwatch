package alert_test

import (
	"testing"
	"time"

	"github.com/yourorg/cronwatch/internal/alert"
)

// captureNotifier records alerts for test assertions.
type captureNotifier struct {
	Alerts []alert.Alert
}

func (c *captureNotifier) Send(a alert.Alert) error {
	c.Alerts = append(c.Alerts, a)
	return nil
}

func TestNewManager_DefaultsToLogNotifier(t *testing.T) {
	m := alert.NewManager()
	if m == nil {
		t.Fatal("expected non-nil Manager")
	}
}

func TestWarn_DispatchesAlert(t *testing.T) {
	cap := &captureNotifier{}
	m := alert.NewManager(cap)

	m.Warn("backup-job", "drift of %d seconds detected", 45)

	if len(cap.Alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(cap.Alerts))
	}
	a := cap.Alerts[0]
	if a.Level != alert.LevelWarn {
		t.Errorf("expected WARN, got %s", a.Level)
	}
	if a.JobName != "backup-job" {
		t.Errorf("unexpected job name: %s", a.JobName)
	}
	if a.Message == "" {
		t.Error("expected non-empty message")
	}
	if a.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestError_DispatchesAlert(t *testing.T) {
	cap := &captureNotifier{}
	m := alert.NewManager(cap)

	m.Error("sync-job", "job did not run for %s", time.Minute)

	if len(cap.Alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(cap.Alerts))
	}
	if cap.Alerts[0].Level != alert.LevelError {
		t.Errorf("expected ERROR, got %s", cap.Alerts[0].Level)
	}
}

func TestDispatch_MultipleNotifiers(t *testing.T) {
	c1 := &captureNotifier{}
	c2 := &captureNotifier{}
	m := alert.NewManager(c1, c2)

	m.Warn("job-a", "test alert")

	if len(c1.Alerts) != 1 || len(c2.Alerts) != 1 {
		t.Errorf("expected each notifier to receive 1 alert, got %d and %d",
			len(c1.Alerts), len(c2.Alerts))
	}
}
