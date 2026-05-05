package monitor

import (
	"testing"
	"time"
)

func makeRecord(job string, drift time.Duration, overdue bool) RunRecord {
	return RunRecord{
		JobName:   job,
		StartedAt: time.Now(),
		Drift:     drift,
		Overdue:   overdue,
	}
}

func TestNewHistory_DefaultMaxSize(t *testing.T) {
	h := NewHistory(0)
	if h.maxSize != 50 {
		t.Errorf("expected default maxSize 50, got %d", h.maxSize)
	}
}

func TestHistory_RecordAndGet(t *testing.T) {
	h := NewHistory(10)
	r := makeRecord("backup", 2*time.Second, false)
	h.Record(r)

	records := h.Get("backup")
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].JobName != "backup" {
		t.Errorf("unexpected job name: %s", records[0].JobName)
	}
}

func TestHistory_EvictsOldestWhenFull(t *testing.T) {
	h := NewHistory(3)
	for i := 0; i < 4; i++ {
		h.Record(makeRecord("cleanup", time.Duration(i)*time.Second, false))
	}

	records := h.Get("cleanup")
	if len(records) != 3 {
		t.Fatalf("expected 3 records after eviction, got %d", len(records))
	}
	// Oldest (drift=0s) should have been evicted; first remaining drift=1s
	if records[0].Drift != 1*time.Second {
		t.Errorf("expected oldest evicted, first drift=1s, got %v", records[0].Drift)
	}
}

func TestHistory_GetUnknownJob(t *testing.T) {
	h := NewHistory(10)
	records := h.Get("nonexistent")
	if records == nil || len(records) != 0 {
		t.Errorf("expected empty slice for unknown job, got %v", records)
	}
}

func TestHistory_All(t *testing.T) {
	h := NewHistory(10)
	h.Record(makeRecord("jobA", 0, false))
	h.Record(makeRecord("jobB", 5*time.Second, true))

	all := h.All()
	if len(all) != 2 {
		t.Errorf("expected 2 jobs in All(), got %d", len(all))
	}
}

func TestHistory_Clear(t *testing.T) {
	h := NewHistory(10)
	h.Record(makeRecord("jobA", 0, false))
	h.Clear("jobA")

	if len(h.Get("jobA")) != 0 {
		t.Error("expected empty records after Clear")
	}
}

func TestHistory_IsolatedCopies(t *testing.T) {
	h := NewHistory(10)
	h.Record(makeRecord("jobA", 0, false))

	copy1 := h.Get("jobA")
	copy1[0].JobName = "mutated"

	copy2 := h.Get("jobA")
	if copy2[0].JobName != "jobA" {
		t.Error("Get should return an isolated copy; internal state was mutated")
	}
}
