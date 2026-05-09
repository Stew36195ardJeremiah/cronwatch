package monitor

import (
	"testing"
	"time"
)

func TestNewNoteStore_Empty(t *testing.T) {
	s := NewNoteStore()
	if len(s.All()) != 0 {
		t.Fatal("expected empty store")
	}
}

func TestNoteStore_SetAndGet(t *testing.T) {
	s := NewNoteStore()
	before := time.Now()
	s.Set("backup", "investigated — false alarm", "alice")
	after := time.Now()

	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected note to exist")
	}
	if e.Job != "backup" {
		t.Errorf("job mismatch: got %q", e.Job)
	}
	if e.Note != "investigated — false alarm" {
		t.Errorf("note mismatch: got %q", e.Note)
	}
	if e.Author != "alice" {
		t.Errorf("author mismatch: got %q", e.Author)
	}
	if e.CreatedAt.Before(before) || e.CreatedAt.After(after) {
		t.Error("created_at out of expected range")
	}
}

func TestNoteStore_GetUnknownJob(t *testing.T) {
	s := NewNoteStore()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected no note for unknown job")
	}
}

func TestNoteStore_SetOverwritesPrevious(t *testing.T) {
	s := NewNoteStore()
	s.Set("deploy", "first note", "alice")
	s.Set("deploy", "updated note", "bob")

	e, ok := s.Get("deploy")
	if !ok {
		t.Fatal("expected note")
	}
	if e.Note != "updated note" {
		t.Errorf("expected updated note, got %q", e.Note)
	}
	if e.Author != "bob" {
		t.Errorf("expected author bob, got %q", e.Author)
	}
}

func TestNoteStore_Delete(t *testing.T) {
	s := NewNoteStore()
	s.Set("cleanup", "some note", "")
	s.Delete("cleanup")
	_, ok := s.Get("cleanup")
	if ok {
		t.Fatal("expected note to be deleted")
	}
}

func TestNoteStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewNoteStore()
	s.Set("job-a", "note a", "alice")
	s.Set("job-b", "note b", "bob")

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 notes, got %d", len(all))
	}
}
