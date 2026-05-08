package monitor

import (
	"testing"
	"time"
)

func TestNewAcknowledgementStore_Empty(t *testing.T) {
	s := NewAcknowledgementStore()
	if len(s.All()) != 0 {
		t.Fatal("expected empty store")
	}
}

func TestAcknowledgementStore_AcknowledgeAndIsAcknowledged(t *testing.T) {
	s := NewAcknowledgementStore()
	s.Acknowledge("job1", "alice", "planned maintenance", time.Hour)
	if !s.IsAcknowledged("job1") {
		t.Fatal("expected job1 to be acknowledged")
	}
}

func TestAcknowledgementStore_UnknownJobNotAcknowledged(t *testing.T) {
	s := NewAcknowledgementStore()
	if s.IsAcknowledged("unknown") {
		t.Fatal("expected unknown job to not be acknowledged")
	}
}

func TestAcknowledgementStore_ExpiredIsNotAcknowledged(t *testing.T) {
	s := NewAcknowledgementStore()
	s.Acknowledge("job1", "bob", "brief ack", -time.Second)
	if s.IsAcknowledged("job1") {
		t.Fatal("expected expired acknowledgement to not be active")
	}
}

func TestAcknowledgementStore_Get_ReturnsEntry(t *testing.T) {
	s := NewAcknowledgementStore()
	s.Acknowledge("job2", "carol", "note", time.Hour)
	e, ok := s.Get("job2")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.AckedBy != "carol" {
		t.Errorf("expected AckedBy=carol, got %s", e.AckedBy)
	}
	if e.Note != "note" {
		t.Errorf("expected Note=note, got %s", e.Note)
	}
}

func TestAcknowledgementStore_Get_UnknownJob(t *testing.T) {
	s := NewAcknowledgementStore()
	_, ok := s.Get("nope")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestAcknowledgementStore_Lift_RemovesEntry(t *testing.T) {
	s := NewAcknowledgementStore()
	s.Acknowledge("job3", "dave", "", time.Hour)
	s.Lift("job3")
	if s.IsAcknowledged("job3") {
		t.Fatal("expected job3 to no longer be acknowledged after lift")
	}
}

func TestAcknowledgementStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewAcknowledgementStore()
	s.Acknowledge("a", "x", "", time.Hour)
	s.Acknowledge("b", "y", "", time.Hour)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}
