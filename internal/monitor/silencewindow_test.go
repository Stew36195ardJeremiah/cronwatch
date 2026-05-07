package monitor

import (
	"testing"
	"time"
)

func TestNewSilenceWindowStore_Empty(t *testing.T) {
	s := NewSilenceWindowStore()
	if s.IsSilenced("job-a") {
		t.Fatal("expected not silenced for unknown job")
	}
}

func TestSilenceWindowStore_ActiveWindow_Silences(t *testing.T) {
	s := NewSilenceWindowStore()
	now := time.Now()
	s.Add(SilenceWindow{
		JobName: "job-a",
		Start:   now.Add(-time.Minute),
		End:     now.Add(time.Hour),
		Reason:  "maintenance",
	})
	if !s.IsSilenced("job-a") {
		t.Fatal("expected job-a to be silenced")
	}
}

func TestSilenceWindowStore_ExpiredWindow_NotSilenced(t *testing.T) {
	s := NewSilenceWindowStore()
	now := time.Now()
	s.Add(SilenceWindow{
		JobName: "job-b",
		Start:   now.Add(-2 * time.Hour),
		End:     now.Add(-time.Minute),
		Reason:  "old",
	})
	if s.IsSilenced("job-b") {
		t.Fatal("expected job-b not to be silenced after window expired")
	}
}

func TestSilenceWindowStore_FutureWindow_NotSilenced(t *testing.T) {
	s := NewSilenceWindowStore()
	now := time.Now()
	s.Add(SilenceWindow{
		JobName: "job-c",
		Start:   now.Add(time.Hour),
		End:     now.Add(2 * time.Hour),
	})
	if s.IsSilenced("job-c") {
		t.Fatal("expected job-c not to be silenced before window starts")
	}
}

func TestSilenceWindowStore_Prune_RemovesExpired(t *testing.T) {
	s := NewSilenceWindowStore()
	now := time.Now()
	s.Add(SilenceWindow{JobName: "job-d", Start: now.Add(-2 * time.Hour), End: now.Add(-time.Minute)})
	s.Add(SilenceWindow{JobName: "job-d", Start: now.Add(-time.Minute), End: now.Add(time.Hour)})
	s.Prune()
	all := s.All()
	if len(all["job-d"]) != 1 {
		t.Fatalf("expected 1 window after prune, got %d", len(all["job-d"]))
	}
}

func TestSilenceWindowStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewSilenceWindowStore()
	now := time.Now()
	s.Add(SilenceWindow{JobName: "job-e", Start: now, End: now.Add(time.Hour)})
	s.Add(SilenceWindow{JobName: "job-f", Start: now, End: now.Add(time.Hour)})
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 jobs in snapshot, got %d", len(all))
	}
}
