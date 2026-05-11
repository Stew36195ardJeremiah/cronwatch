package monitor

import (
	"testing"
	"time"
)

func TestNewLockoutStore_Defaults(t *testing.T) {
	s := NewLockoutStore(0, 0)
	if s.threshold != 5 {
		t.Errorf("expected default threshold 5, got %d", s.threshold)
	}
	if s.duration != 15*time.Minute {
		t.Errorf("expected default duration 15m, got %v", s.duration)
	}
}

func TestLockoutStore_NotLockedInitially(t *testing.T) {
	s := NewLockoutStore(3, time.Minute)
	if s.IsLockedOut("job-a") {
		t.Error("expected job-a to not be locked out initially")
	}
}

func TestLockoutStore_LocksAfterThreshold(t *testing.T) {
	s := NewLockoutStore(3, time.Minute)
	for i := 0; i < 3; i++ {
		s.RecordFailure("job-a", "timeout")
	}
	if !s.IsLockedOut("job-a") {
		t.Error("expected job-a to be locked out after threshold")
	}
}

func TestLockoutStore_BelowThresholdNotLocked(t *testing.T) {
	s := NewLockoutStore(5, time.Minute)
	s.RecordFailure("job-b", "exit code 1")
	s.RecordFailure("job-b", "exit code 1")
	if s.IsLockedOut("job-b") {
		t.Error("expected job-b to not be locked out below threshold")
	}
}

func TestLockoutStore_LockExpires(t *testing.T) {
	s := NewLockoutStore(1, 10*time.Millisecond)
	s.RecordFailure("job-c", "oom")
	if !s.IsLockedOut("job-c") {
		t.Fatal("expected job-c to be locked out")
	}
	time.Sleep(20 * time.Millisecond)
	if s.IsLockedOut("job-c") {
		t.Error("expected lockout to have expired")
	}
}

func TestLockoutStore_Lift_RemovesLockout(t *testing.T) {
	s := NewLockoutStore(2, time.Minute)
	s.RecordFailure("job-d", "crash")
	s.RecordFailure("job-d", "crash")
	if !s.IsLockedOut("job-d") {
		t.Fatal("expected job-d to be locked out")
	}
	s.Lift("job-d")
	if s.IsLockedOut("job-d") {
		t.Error("expected lockout to be lifted")
	}
}

func TestLockoutStore_Get_ReturnsEntry(t *testing.T) {
	s := NewLockoutStore(1, time.Hour)
	s.RecordFailure("job-e", "disk full")
	e := s.Get("job-e")
	if e == nil {
		t.Fatal("expected entry, got nil")
	}
	if e.Job != "job-e" {
		t.Errorf("expected job-e, got %s", e.Job)
	}
	if e.Reason != "disk full" {
		t.Errorf("expected reason 'disk full', got %s", e.Reason)
	}
	if e.Trips != 1 {
		t.Errorf("expected 1 trip, got %d", e.Trips)
	}
}

func TestLockoutStore_Get_UnknownJob(t *testing.T) {
	s := NewLockoutStore(3, time.Minute)
	if s.Get("nope") != nil {
		t.Error("expected nil for unknown job")
	}
}

func TestLockoutStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewLockoutStore(1, time.Minute)
	s.RecordFailure("j1", "err")
	s.RecordFailure("j2", "err")
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestLockoutStore_TripsIncrement(t *testing.T) {
	s := NewLockoutStore(1, 10*time.Millisecond)
	s.RecordFailure("job-f", "err")
	time.Sleep(20 * time.Millisecond)
	s.RecordFailure("job-f", "err")
	e := s.Get("job-f")
	if e == nil {
		t.Fatal("expected entry")
	}
	if e.Trips != 2 {
		t.Errorf("expected 2 trips, got %d", e.Trips)
	}
}
