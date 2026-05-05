package monitor

import (
	"testing"
	"time"

	"github.com/user/cronwatch/internal/config"
)

func newCheckerConfig() *config.Config {
	return &config.Config{
		Jobs: []config.Job{
			{
				Name:     "test-job",
				Schedule: "* * * * *",
				Timeout:  "5m",
			},
			{
				Name:     "hourly-job",
				Schedule: "0 * * * *",
				Timeout:  "10m",
			},
		},
	}
}

func TestChecker_CheckAll_NoLastRun(t *testing.T) {
	cfg := newCheckerConfig()
	store := NewStatusStore()
	hist := NewHistory(10)
	alerts := newMockAlertManager()

	checker := newChecker(cfg, store, hist, alerts)
	checker.checkAll()

	statuses := store.All()
	if len(statuses) == 0 {
		t.Fatal("expected statuses to be populated after checkAll")
	}
}

func TestChecker_CheckAll_OverdueJob(t *testing.T) {
	cfg := &config.Config{
		Jobs: []config.Job{
			{
				Name:     "stale-job",
				Schedule: "* * * * *",
				Timeout:  "1m",
			},
		},
	}
	store := NewStatusStore()
	hist := NewHistory(10)
	alerts := newMockAlertManager()

	// Simulate a last run far in the past
	store.Set("stale-job", JobStatus{
		Name:    "stale-job",
		LastRun: time.Now().Add(-10 * time.Minute),
		OK:      true,
	})

	checker := newChecker(cfg, store, hist, alerts)
	checker.checkAll()

	status, ok := store.Get("stale-job")
	if !ok {
		t.Fatal("expected status for stale-job")
	}
	if status.OK {
		t.Error("expected stale-job to be marked not OK due to overdue")
	}
	if alerts.warnCount == 0 && alerts.errorCount == 0 {
		t.Error("expected at least one alert to be dispatched for overdue job")
	}
}

func TestChecker_CheckAll_RecentRun(t *testing.T) {
	cfg := &config.Config{
		Jobs: []config.Job{
			{
				Name:     "fresh-job",
				Schedule: "* * * * *",
				Timeout:  "5m",
			},
		},
	}
	store := NewStatusStore()
	hist := NewHistory(10)
	alerts := newMockAlertManager()

	store.Set("fresh-job", JobStatus{
		Name:    "fresh-job",
		LastRun: time.Now().Add(-30 * time.Second),
		OK:      true,
	})

	checker := newChecker(cfg, store, hist, alerts)
	checker.checkAll()

	status, ok := store.Get("fresh-job")
	if !ok {
		t.Fatal("expected status for fresh-job")
	}
	if !status.OK {
		t.Error("expected fresh-job to remain OK")
	}
}

// mockAlertManager satisfies the alertManager interface used in checker.go
type mockAlertManager struct {
	warnCount  int
	errorCount int
}

func newMockAlertManager() *mockAlertManager {
	return &mockAlertManager{}
}

func (m *mockAlertManager) Warn(job, msg string) {
	m.warnCount++
}

func (m *mockAlertManager) Error(job, msg string) {
	m.errorCount++
}
