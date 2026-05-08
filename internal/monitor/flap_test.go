package monitor

import (
	"testing"
	"time"
)

func TestNewFlapStore_Defaults(t *testing.T) {
	f := NewFlapStore(0, 0)
	if f.window != 10*time.Minute {
		t.Errorf("expected default window 10m, got %v", f.window)
	}
	if f.threshold != 4 {
		t.Errorf("expected default threshold 4, got %d", f.threshold)
	}
}

func TestFlapStore_NotFlappingInitially(t *testing.T) {
	f := NewFlapStore(10*time.Minute, 3)
	if f.IsFlapping("job1") {
		t.Error("expected not flapping for unknown job")
	}
}

func TestFlapStore_FlapsAfterThreshold(t *testing.T) {
	f := NewFlapStore(10*time.Minute, 3)
	now := time.Now()

	f.RecordTransition("job1", now.Add(-3*time.Minute))
	f.RecordTransition("job1", now.Add(-2*time.Minute))
	if f.IsFlapping("job1") {
		t.Error("should not be flapping before threshold")
	}
	f.RecordTransition("job1", now.Add(-1*time.Minute))
	if !f.IsFlapping("job1") {
		t.Error("expected flapping after threshold transitions")
	}
}

func TestFlapStore_PrunesOldTransitions(t *testing.T) {
	f := NewFlapStore(5*time.Minute, 3)
	now := time.Now()

	// These are outside the window.
	f.RecordTransition("job1", now.Add(-10*time.Minute))
	f.RecordTransition("job1", now.Add(-8*time.Minute))
	f.RecordTransition("job1", now.Add(-6*time.Minute))

	// Only one recent transition — should not flap.
	flapping := f.RecordTransition("job1", now)
	if flapping {
		t.Error("expected not flapping after old transitions pruned")
	}
}

func TestFlapStore_Reset_ClearsState(t *testing.T) {
	f := NewFlapStore(10*time.Minute, 2)
	now := time.Now()
	f.RecordTransition("job1", now.Add(-2*time.Minute))
	f.RecordTransition("job1", now)
	if !f.IsFlapping("job1") {
		t.Fatal("expected flapping before reset")
	}
	f.Reset("job1")
	if f.IsFlapping("job1") {
		t.Error("expected not flapping after reset")
	}
	if f.Get("job1") != nil {
		t.Error("expected nil entry after reset")
	}
}

func TestFlapStore_All_ReturnsSnapshot(t *testing.T) {
	f := NewFlapStore(10*time.Minute, 2)
	now := time.Now()
	f.RecordTransition("jobA", now.Add(-1*time.Minute))
	f.RecordTransition("jobA", now)
	f.RecordTransition("jobB", now)

	all := f.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
	if !all["jobA"].IsFlapping {
		t.Error("expected jobA to be flapping")
	}
	if all["jobB"].IsFlapping {
		t.Error("expected jobB not to be flapping")
	}
}
