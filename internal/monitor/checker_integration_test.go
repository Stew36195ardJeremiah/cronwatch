package monitor

import (
	"testing"
	"time"

	"github.com/user/cronwatch/internal/config"
)

func TestChecker_Integration_OverdueTriggersAlert(t *testing.T) {
	cfg := &config.Config{
		Jobs: []config.Job{
			{
				Name:     "integration-job",
				Schedule: "* * * * *",
				Timeout:  "1m",
			},
		},
	}

	store := NewStatusStore()
	hist := NewHistory(50)
	alerts := newMockAlertManager()

	// Mark job as last run 15 minutes ago — clearly overdue for a per-minute schedule
	store.Set("integration-job", JobStatus{
		Name:    "integration-job",
		LastRun: time.Now().Add(-15 * time.Minute),
		OK:      true,
	})

	checker := newChecker(cfg, store, hist, alerts)
	checker.checkAll()

	if alerts.warnCount+alerts.errorCount == 0 {
		t.Error("expected alert to be fired for overdue integration-job")
	}

	status, ok := store.Get("integration-job")
	if !ok {
		t.Fatal("status should exist after check")
	}
	if status.OK {
		t.Error("status should be NOT OK for overdue job")
	}
}

func TestChecker_Integration_HistoryRecordedOnOverdue(t *testing.T) {
	cfg := &config.Config{
		Jobs: []config.Job{
			{
				Name:     "hist-job",
				Schedule: "* * * * *",
				Timeout:  "1m",
			},
		},
	}

	store := NewStatusStore()
	hist := NewHistory(50)
	alerts := newMockAlertManager()

	store.Set("hist-job", JobStatus{
		Name:    "hist-job",
		LastRun: time.Now().Add(-20 * time.Minute),
		OK:      true,
	})

	checker := newChecker(cfg, store, hist, alerts)
	checker.checkAll()

	records := hist.Get("hist-job")
	if len(records) == 0 {
		t.Error("expected history to be recorded for overdue hist-job")
	}
}
