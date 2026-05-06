package monitor

import (
	"testing"
	"time"
)

func makeAnnotation(job, author, note string) Annotation {
	return Annotation{
		JobName:   job,
		Author:    author,
		Note:      note,
		CreatedAt: time.Now().UTC(),
	}
}

func TestNewAnnotationStore_DefaultMaxSize(t *testing.T) {
	s := NewAnnotationStore(0)
	if s.maxPer != 50 {
		t.Fatalf("expected default maxPer=50, got %d", s.maxPer)
	}
}

func TestAnnotationStore_AddAndGet(t *testing.T) {
	s := NewAnnotationStore(10)
	s.Add(makeAnnotation("backup", "alice", "scheduled maintenance"))
	s.Add(makeAnnotation("backup", "bob", "post-deploy check"))

	list := s.Get("backup")
	if len(list) != 2 {
		t.Fatalf("expected 2 annotations, got %d", len(list))
	}
	if list[0].Author != "alice" {
		t.Errorf("expected first author alice, got %s", list[0].Author)
	}
}

func TestAnnotationStore_GetUnknownJob(t *testing.T) {
	s := NewAnnotationStore(10)
	list := s.Get("nonexistent")
	if len(list) != 0 {
		t.Errorf("expected empty slice for unknown job")
	}
}

func TestAnnotationStore_EvictsOldestWhenFull(t *testing.T) {
	s := NewAnnotationStore(3)
	for i := 0; i < 4; i++ {
		s.Add(makeAnnotation("job", "user", string(rune('a'+i))))
	}
	list := s.Get("job")
	if len(list) != 3 {
		t.Fatalf("expected 3 after eviction, got %d", len(list))
	}
	if list[0].Note != "b" {
		t.Errorf("expected oldest evicted, first note should be 'b', got %s", list[0].Note)
	}
}

func TestAnnotationStore_Delete(t *testing.T) {
	s := NewAnnotationStore(10)
	s.Add(makeAnnotation("cleanup", "ops", "note"))
	s.Delete("cleanup")
	if len(s.Get("cleanup")) != 0 {
		t.Error("expected empty after delete")
	}
}

func TestAnnotationStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewAnnotationStore(10)
	s.Add(makeAnnotation("job-a", "u", "n1"))
	s.Add(makeAnnotation("job-b", "u", "n2"))
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 total annotations, got %d", len(all))
	}
}

func TestAnnotationStore_SetsCreatedAt(t *testing.T) {
	s := NewAnnotationStore(10)
	a := Annotation{JobName: "job", Author: "x", Note: "y"}
	before := time.Now().UTC()
	s.Add(a)
	after := time.Now().UTC()
	list := s.Get("job")
	if list[0].CreatedAt.Before(before) || list[0].CreatedAt.After(after) {
		t.Error("CreatedAt not set to current time")
	}
}
