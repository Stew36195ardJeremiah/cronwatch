package monitor

import (
	"testing"
	"time"
)

func TestNewFingerprintStore_Defaults(t *testing.T) {
	fs := NewFingerprintStore(0)
	if fs.ttl != 24*time.Hour {
		t.Errorf("expected default ttl 24h, got %v", fs.ttl)
	}
}

func TestFingerprintStore_NewEntryIsNew(t *testing.T) {
	fs := NewFingerprintStore(time.Minute)
	isNew := fs.Record("job1", "warn", "overdue")
	if !isNew {
		t.Error("expected first record to be new")
	}
}

func TestFingerprintStore_DuplicateIsNotNew(t *testing.T) {
	fs := NewFingerprintStore(time.Minute)
	fs.Record("job1", "warn", "overdue")
	isNew := fs.Record("job1", "warn", "overdue")
	if isNew {
		t.Error("expected duplicate record to not be new")
	}
}

func TestFingerprintStore_DifferentLevelIsNew(t *testing.T) {
	fs := NewFingerprintStore(time.Minute)
	fs.Record("job1", "warn", "overdue")
	isNew := fs.Record("job1", "error", "overdue")
	if !isNew {
		t.Error("expected different level to produce new fingerprint")
	}
}

func TestFingerprintStore_CountIncrements(t *testing.T) {
	fs := NewFingerprintStore(time.Minute)
	fs.Record("job1", "warn", "overdue")
	fs.Record("job1", "warn", "overdue")
	fs.Record("job1", "warn", "overdue")

	all := fs.All()
	if len(all) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(all))
	}
	if all[0].Count != 3 {
		t.Errorf("expected count 3, got %d", all[0].Count)
	}
}

func TestFingerprintStore_All_ReturnsSnapshot(t *testing.T) {
	fs := NewFingerprintStore(time.Minute)
	fs.Record("jobA", "warn", "msg")
	fs.Record("jobB", "error", "msg")

	all := fs.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}

func TestFingerprintStore_ExpiredEntryBecomesNew(t *testing.T) {
	fs := NewFingerprintStore(1 * time.Millisecond)
	fs.Record("job1", "warn", "overdue")

	time.Sleep(5 * time.Millisecond)

	// Next record should evict the old entry and treat this as new.
	isNew := fs.Record("job1", "warn", "overdue")
	if !isNew {
		t.Error("expected expired entry to be treated as new")
	}
}

func TestFingerprintStore_HashConsistency(t *testing.T) {
	h1 := fingerprintHash("job", "warn", "msg")
	h2 := fingerprintHash("job", "warn", "msg")
	if h1 != h2 {
		t.Errorf("expected consistent hashes, got %s and %s", h1, h2)
	}
}
