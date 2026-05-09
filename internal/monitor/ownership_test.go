package monitor

import (
	"testing"
	"time"
)

func TestNewOwnershipStore_Empty(t *testing.T) {
	s := NewOwnershipStore()
	all := s.All()
	if len(all) != 0 {
		t.Fatalf("expected empty store, got %d entries", len(all))
	}
}

func TestOwnershipStore_SetAndGet(t *testing.T) {
	s := NewOwnershipStore()
	s.Set("backup", "alice", "infra", "alice@example.com")

	e, ok := s.Get("backup")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.Owner != "alice" {
		t.Errorf("expected owner alice, got %s", e.Owner)
	}
	if e.Team != "infra" {
		t.Errorf("expected team infra, got %s", e.Team)
	}
	if e.Contact != "alice@example.com" {
		t.Errorf("expected contact alice@example.com, got %s", e.Contact)
	}
	if e.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestOwnershipStore_GetUnknownJob(t *testing.T) {
	s := NewOwnershipStore()
	_, ok := s.Get("nonexistent")
	if ok {
		t.Fatal("expected no entry for unknown job")
	}
}

func TestOwnershipStore_Remove(t *testing.T) {
	s := NewOwnershipStore()
	s.Set("deploy", "bob", "platform", "bob@example.com")
	s.Remove("deploy")
	_, ok := s.Get("deploy")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestOwnershipStore_All_ReturnsSnapshot(t *testing.T) {
	s := NewOwnershipStore()
	s.Set("job-a", "alice", "infra", "alice@example.com")
	s.Set("job-b", "bob", "platform", "bob@example.com")

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
}

func TestOwnershipStore_OverwriteUpdatesTimestamp(t *testing.T) {
	s := NewOwnershipStore()
	s.Set("job-a", "alice", "infra", "alice@example.com")

	first, _ := s.Get("job-a")
	time.Sleep(2 * time.Millisecond)
	s.Set("job-a", "carol", "sre", "carol@example.com")

	second, _ := s.Get("job-a")
	if !second.UpdatedAt.After(first.UpdatedAt) {
		t.Error("expected UpdatedAt to advance on overwrite")
	}
	if second.Owner != "carol" {
		t.Errorf("expected owner carol, got %s", second.Owner)
	}
}
