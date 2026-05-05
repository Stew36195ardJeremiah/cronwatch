package monitor_test

import (
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/monitor"
)

// TestHistory_RecordedAfterCheckerRun verifies that the Monitor populates
// History when RecordRun is called, simulating the checker integration path.
func TestHistory_RecordedAfterCheckerRun(t *testing.T) {
	cfg := testConfig()
	m := monitor.New(cfg)

	// Simulate the checker recording a run for the first job.
	jobName := cfg.Jobs[0].Name
	now := time.Now()
	m.RecordRun(jobName, now)

	// History should reflect this run.
	records := m.History.Get(jobName)
	if len(records) == 0 {
		t.Fatal("expected at least one history record after RecordRun")
	}

	if records[0].JobName != jobName {
		t.Errorf("history record job name mismatch: got %q, want %q", records[0].JobName, jobName)
	}
}

// TestHistory_MultipleJobsTrackedIndependently ensures separate job histories
// do not bleed into each other.
func TestHistory_MultipleJobsTrackedIndependently(t *testing.T) {
	cfg := testConfig()
	// Add a second job by duplicating the first with a different name.
	secondJob := cfg.Jobs[0]
	secondJob.Name = "second-job"
	cfg.Jobs = append(cfg.Jobs, secondJob)

	m := monitor.New(cfg)
	now := time.Now()

	m.RecordRun(cfg.Jobs[0].Name, now)
	m.RecordRun(cfg.Jobs[0].Name, now.Add(time.Minute))
	m.RecordRun("second-job", now)

	if len(m.History.Get(cfg.Jobs[0].Name)) != 2 {
		t.Errorf("expected 2 records for first job")
	}
	if len(m.History.Get("second-job")) != 1 {
		t.Errorf("expected 1 record for second job")
	}
}
