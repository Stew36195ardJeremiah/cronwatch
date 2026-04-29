package monitor_test

import (
	"testing"
	"time"

	"github.com/cronwatch/internal/config"
	"github.com/cronwatch/internal/monitor"
)

func testConfig(schedule string) *config.Config {
	return &config.Config{
		Jobs: []config.Job{
			{Name: "test-job", Schedule: schedule},
		},
	}
}

func TestNew_InitializesMonitor(t *testing.T) {
	cfg := testConfig("* * * * *")
	m := monitor.New(cfg)
	if m == nil {
		t.Fatal("expected non-nil monitor")
	}
}

func TestStatuses_EmptyBeforeCheck(t *testing.T) {
	cfg := testConfig("* * * * *")
	m := monitor.New(cfg)
	statuses := m.Statuses()
	if len(statuses) != 0 {
		t.Errorf("expected 0 statuses before any check, got %d", len(statuses))
	}
}

func TestRecordRun_UpdatesLastRun(t *testing.T) {
	cfg := testConfig("* * * * *")
	m := monitor.New(cfg)

	// Trigger internal state initialization via Start + brief wait.
	m.Start(10 * time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	m.Stop()

	now := time.Now()
	m.RecordRun("test-job", now)

	statuses := m.Statuses()
	if len(statuses) == 0 {
		t.Fatal("expected at least one status after polling")
	}
	for _, s := range statuses {
		if s.Name == "test-job" {
			if s.LastRun.IsZero() {
				t.Error("expected LastRun to be set after RecordRun")
			}
			if s.Overdue {
				t.Error("expected job not to be overdue after RecordRun")
			}
			return
		}
	}
	t.Error("test-job not found in statuses")
}

func TestStartStop_DoesNotPanic(t *testing.T) {
	cfg := testConfig("*/5 * * * *")
	m := monitor.New(cfg)
	m.Start(50 * time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	m.Stop()
}
