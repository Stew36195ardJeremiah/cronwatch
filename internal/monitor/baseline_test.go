package monitor

import (
	"testing"
	"time"
)

func TestNewBaselineStore_Defaults(t *testing.T) {
	s := NewBaselineStore(0)
	if s.maxSamples != 10 {
		t.Fatalf("expected default maxSamples=10, got %d", s.maxSamples)
	}
}

func TestBaselineStore_RecordAndGet(t *testing.T) {
	s := NewBaselineStore(5)
	s.Record("job1", 2*time.Second)
	s.Record("job1", 4*time.Second)

	e, ok := s.Get("job1")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.SampleCount != 2 {
		t.Fatalf("expected SampleCount=2, got %d", e.SampleCount)
	}
	if e.AvgDuration != 3*time.Second {
		t.Fatalf("expected avg=3s, got %s", e.AvgDuration)
	}
}

func TestBaselineStore_GetUnknownJob(t *testing.T) {
	s := NewBaselineStore(5)
	_, ok := s.Get("ghost")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestBaselineStore_CapsAtMaxSamples(t *testing.T) {
	s := NewBaselineStore(3)
	for i := 0; i < 10; i++ {
		s.Record("job", time.Second)
	}
	e, _ := s.Get("job")
	if e.SampleCount > 3 {
		t.Fatalf("SampleCount should not exceed maxSamples, got %d", e.SampleCount)
	}
}

func TestBaselineStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewBaselineStore(5)
	s.Record("a", time.Second)
	s.Record("b", 2*time.Second)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestBaselineStore_Reset_ClearsEntry(t *testing.T) {
	s := NewBaselineStore(5)
	s.Record("job", time.Second)
	s.Reset("job")
	_, ok := s.Get("job")
	if ok {
		t.Fatal("expected entry to be cleared after reset")
	}
}

func TestBaselineStore_LastUpdated_IsSet(t *testing.T) {
	s := NewBaselineStore(5)
	before := time.Now()
	s.Record("job", time.Second)
	e, _ := s.Get("job")
	if e.LastUpdated.Before(before) {
		t.Fatal("expected LastUpdated to be recent")
	}
}
