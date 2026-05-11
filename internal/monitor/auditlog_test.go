package monitor

import (
	"testing"
	"time"
)

func TestNewAuditLogStore_Empty(t *testing.T) {
	s := NewAuditLogStore()
	if got := s.All(); len(got) != 0 {
		t.Fatalf("expected empty store, got %d entries", len(got))
	}
}

func TestAuditLogStore_RecordAndAll(t *testing.T) {
	s := NewAuditLogStore()
	s.Record("job1", AuditActionPause, "admin", "maintenance")
	s.Record("job2", AuditActionSuppress, "ci", "deploy")

	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all[0].Job != "job1" || all[0].Action != AuditActionPause {
		t.Errorf("unexpected first entry: %+v", all[0])
	}
	if all[1].Actor != "ci" {
		t.Errorf("expected actor 'ci', got %q", all[1].Actor)
	}
}

func TestAuditLogStore_TimestampIsSet(t *testing.T) {
	s := NewAuditLogStore()
	before := time.Now().UTC()
	s.Record("job1", AuditActionResume, "user", "")
	after := time.Now().UTC()

	all := s.All()
	if all[0].Timestamp.Before(before) || all[0].Timestamp.After(after) {
		t.Errorf("timestamp out of expected range: %v", all[0].Timestamp)
	}
}

func TestAuditLogStore_EvictsOldestWhenFull(t *testing.T) {
	s := &AuditLogStore{maxSize: 3}
	s.Record("a", AuditActionPause, "u", "first")
	s.Record("b", AuditActionPause, "u", "second")
	s.Record("c", AuditActionPause, "u", "third")
	s.Record("d", AuditActionPause, "u", "fourth")

	all := s.All()
	if len(all) != 3 {
		t.Fatalf("expected 3 entries after eviction, got %d", len(all))
	}
	if all[0].Detail != "second" {
		t.Errorf("expected oldest evicted, first entry detail = %q", all[0].Detail)
	}
}

func TestAuditLogStore_ForJob_FiltersCorrectly(t *testing.T) {
	s := NewAuditLogStore()
	s.Record("alpha", AuditActionSuppress, "u", "")
	s.Record("beta", AuditActionLift, "u", "")
	s.Record("alpha", AuditActionResume, "u", "")

	got := s.ForJob("alpha")
	if len(got) != 2 {
		t.Fatalf("expected 2 entries for alpha, got %d", len(got))
	}
	for _, e := range got {
		if e.Job != "alpha" {
			t.Errorf("unexpected job in filtered result: %q", e.Job)
		}
	}
}

func TestAuditLogStore_ForJob_UnknownJob(t *testing.T) {
	s := NewAuditLogStore()
	s.Record("known", AuditActionPause, "u", "")
	if got := s.ForJob("unknown"); len(got) != 0 {
		t.Errorf("expected empty slice for unknown job, got %d", len(got))
	}
}
