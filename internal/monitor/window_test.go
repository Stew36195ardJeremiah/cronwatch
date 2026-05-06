package monitor

import (
	"testing"
	"time"
)

func TestNewWindowStore_Defaults(t *testing.T) {
	w := NewWindowStore(0)
	if w.window != time.Hour {
		t.Fatalf("expected default window 1h, got %v", w.window)
	}
}

func TestWindowStore_RecordAndCount(t *testing.T) {
	now := time.Now()
	w := NewWindowStore(time.Hour)
	w.Record("backup", now)
	w.Record("backup", now.Add(10*time.Second))
	w.Record("backup", now.Add(2*time.Minute))

	count := w.Count("backup", now.Add(3*time.Minute))
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}
}

func TestWindowStore_CountUnknownJob(t *testing.T) {
	w := NewWindowStore(time.Hour)
	count := w.Count("ghost", time.Now())
	if count != 0 {
		t.Fatalf("expected 0 for unknown job, got %d", count)
	}
}

func TestWindowStore_EvictsOldEntries(t *testing.T) {
	w := NewWindowStore(30 * time.Minute)
	base := time.Now()

	// Record two events well outside the window.
	w.Record("sync", base.Add(-60*time.Minute))
	w.Record("sync", base.Add(-45*time.Minute))
	// Record one inside the window.
	w.Record("sync", base.Add(-10*time.Minute))

	count := w.Count("sync", base)
	if count != 1 {
		t.Fatalf("expected 1 after eviction, got %d", count)
	}
}

func TestWindowStore_AllReturnsSnapshot(t *testing.T) {
	w := NewWindowStore(time.Hour)
	now := time.Now()
	w.Record("jobA", now)
	w.Record("jobB", now)
	w.Record("jobB", now.Add(time.Second))

	entries := w.All(now)
	totals := make(map[string]int)
	for _, e := range entries {
		totals[e.JobName] += e.Count
	}
	if totals["jobA"] != 1 {
		t.Errorf("expected jobA count 1, got %d", totals["jobA"])
	}
	if totals["jobB"] != 2 {
		t.Errorf("expected jobB count 2, got %d", totals["jobB"])
	}
}

func TestWindowStore_MultipleJobsIndependent(t *testing.T) {
	w := NewWindowStore(time.Hour)
	now := time.Now()
	w.Record("alpha", now)
	w.Record("beta", now)
	w.Record("beta", now.Add(time.Minute))

	if c := w.Count("alpha", now.Add(2*time.Minute)); c != 1 {
		t.Errorf("alpha: expected 1, got %d", c)
	}
	if c := w.Count("beta", now.Add(2*time.Minute)); c != 2 {
		t.Errorf("beta: expected 2, got %d", c)
	}
}
