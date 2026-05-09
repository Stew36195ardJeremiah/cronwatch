package monitor

import (
	"testing"
	"time"
)

func TestNewSLAStore_Empty(t *testing.T) {
	s := NewSLAStore()
	if got := s.All(); len(got) != 0 {
		t.Fatalf("expected empty store, got %d entries", len(got))
	}
}

func TestSLAStore_SetAndGet(t *testing.T) {
	s := NewSLAStore()
	s.Set("backup", 5*time.Minute, time.Time{})
	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.MaxDuration != 5*time.Minute {
		t.Errorf("unexpected MaxDuration: %v", e.MaxDuration)
	}
	if !e.Compliant {
		t.Error("new entry should default to compliant")
	}
}

func TestSLAStore_GetUnknownJob(t *testing.T) {
	s := NewSLAStore()
	_, ok := s.Get("unknown")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestSLAStore_RecordRun_Compliant(t *testing.T) {
	s := NewSLAStore()
	s.Set("etl", 10*time.Minute, time.Time{})
	s.RecordRun("etl", 3*time.Minute, time.Now())
	e, _ := s.Get("etl")
	if !e.Compliant {
		t.Error("run within SLA should be compliant")
	}
	if e.ViolationCount != 0 {
		t.Errorf("expected 0 violations, got %d", e.ViolationCount)
	}
}

func TestSLAStore_RecordRun_ViolatesMaxDuration(t *testing.T) {
	s := NewSLAStore()
	s.Set("etl", 2*time.Minute, time.Time{})
	s.RecordRun("etl", 5*time.Minute, time.Now())
	e, _ := s.Get("etl")
	if e.Compliant {
		t.Error("run exceeding max duration should not be compliant")
	}
	if e.ViolationCount != 1 {
		t.Errorf("expected 1 violation, got %d", e.ViolationCount)
	}
}

func TestSLAStore_RecordRun_ViolatesDeadline(t *testing.T) {
	s := NewSLAStore()
	deadline := time.Now().Add(-1 * time.Minute) // already passed
	s.Set("report", 0, deadline)
	s.RecordRun("report", 30*time.Second, time.Now())
	e, _ := s.Get("report")
	if e.Compliant {
		t.Error("run finishing after deadline should not be compliant")
	}
}

func TestSLAStore_RecordRun_UnknownJobIsNoop(t *testing.T) {
	s := NewSLAStore()
	// should not panic
	s.RecordRun("ghost", time.Minute, time.Now())
}

func TestSLAStore_Remove(t *testing.T) {
	s := NewSLAStore()
	s.Set("job1", time.Minute, time.Time{})
	s.Remove("job1")
	_, ok := s.Get("job1")
	if ok {
		t.Error("expected entry to be removed")
	}
}

func TestSLAStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewSLAStore()
	s.Set("a", time.Minute, time.Time{})
	s.Set("b", 2*time.Minute, time.Time{})
	all := s.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}
