package monitor

import (
	"testing"
)

func TestNewPauseStore_Empty(t *testing.T) {
	s := NewPauseStore()
	if len(s.All()) != 0 {
		t.Fatal("expected empty store")
	}
}

func TestPauseStore_PauseAndIsPaused(t *testing.T) {
	s := NewPauseStore()
	s.Pause("backup", "maintenance", "alice")
	if !s.IsPaused("backup") {
		t.Fatal("expected job to be paused")
	}
}

func TestPauseStore_Resume_RemovesPause(t *testing.T) {
	s := NewPauseStore()
	s.Pause("backup", "maintenance", "alice")
	s.Resume("backup")
	if s.IsPaused("backup") {
		t.Fatal("expected job to be resumed")
	}
}

func TestPauseStore_Get_ReturnsEntry(t *testing.T) {
	s := NewPauseStore()
	s.Pause("sync", "deploy", "bob")
	e, ok := s.Get("sync")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Reason != "deploy" || e.PausedBy != "bob" {
		t.Fatalf("unexpected entry: %+v", e)
	}
}

func TestPauseStore_Get_UnknownJob(t *testing.T) {
	s := NewPauseStore()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestPauseStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewPauseStore()
	s.Pause("jobA", "", "")
	s.Pause("jobB", "", "")
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestPauseStore_IsPaused_UnknownReturnsFalse(t *testing.T) {
	s := NewPauseStore()
	if s.IsPaused("ghost") {
		t.Fatal("expected false for unknown job")
	}
}
