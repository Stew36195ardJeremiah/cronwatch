package monitor

import (
	"fmt"
	"testing"
	"time"
)

func makeDeadLetterEntry(job, notifier string) DeadLetterEntry {
	return DeadLetterEntry{
		JobName:  job,
		Level:    "error",
		Message:  "alert dispatch failed",
		Notifier: notifier,
		Err:      "connection refused",
		FailedAt: time.Now(),
		Attempts: 1,
	}
}

func TestNewDeadLetterStore_DefaultMaxSize(t *testing.T) {
	d := NewDeadLetterStore(0)
	if d.maxSize != defaultDeadLetterMax {
		t.Errorf("expected default max %d, got %d", defaultDeadLetterMax, d.maxSize)
	}
}

func TestDeadLetterStore_RecordAndAll(t *testing.T) {
	d := NewDeadLetterStore(10)
	d.Record(makeDeadLetterEntry("job-a", "slack"))
	d.Record(makeDeadLetterEntry("job-b", "email"))

	all := d.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all[0].JobName != "job-a" || all[1].JobName != "job-b" {
		t.Error("entries in wrong order or wrong job names")
	}
}

func TestDeadLetterStore_EvictsOldestWhenFull(t *testing.T) {
	d := NewDeadLetterStore(3)
	for i := 0; i < 4; i++ {
		d.Record(makeDeadLetterEntry(fmt.Sprintf("job-%d", i), "slack"))
	}
	all := d.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(all))
	}
	if all[0].JobName != "job-1" {
		t.Errorf("expected oldest evicted, first entry should be job-1, got %s", all[0].JobName)
	}
}

func TestDeadLetterStore_ForJob_FiltersCorrectly(t *testing.T) {
	d := NewDeadLetterStore(20)
	d.Record(makeDeadLetterEntry("backup", "slack"))
	d.Record(makeDeadLetterEntry("cleanup", "email"))
	d.Record(makeDeadLetterEntry("backup", "pagerduty"))

	results := d.ForJob("backup")
	if len(results) != 2 {
		t.Fatalf("expected 2 entries for 'backup', got %d", len(results))
	}
	for _, r := range results {
		if r.JobName != "backup" {
			t.Errorf("unexpected job name %s in ForJob result", r.JobName)
		}
	}
}

func TestDeadLetterStore_ForJob_UnknownJob(t *testing.T) {
	d := NewDeadLetterStore(10)
	d.Record(makeDeadLetterEntry("job-a", "slack"))
	results := d.ForJob("nonexistent")
	if results != nil && len(results) != 0 {
		t.Errorf("expected empty slice for unknown job, got %v", results)
	}
}

func TestDeadLetterStore_Clear(t *testing.T) {
	d := NewDeadLetterStore(10)
	d.Record(makeDeadLetterEntry("job-a", "slack"))
	d.Record(makeDeadLetterEntry("job-b", "email"))
	d.Clear()
	if d.Len() != 0 {
		t.Errorf("expected 0 entries after Clear, got %d", d.Len())
	}
}

func TestDeadLetterStore_Len(t *testing.T) {
	d := NewDeadLetterStore(10)
	if d.Len() != 0 {
		t.Errorf("expected initial Len 0, got %d", d.Len())
	}
	d.Record(makeDeadLetterEntry("job-a", "slack"))
	if d.Len() != 1 {
		t.Errorf("expected Len 1, got %d", d.Len())
	}
}
