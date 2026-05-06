package monitor

import (
	"testing"
	"time"
)

func TestNewHeartbeatStore_Empty(t *testing.T) {
	h := NewHeartbeatStore()
	if len(h.All()) != 0 {
		t.Fatal("expected empty store")
	}
}

func TestHeartbeatStore_IsExpired_Unknown(t *testing.T) {
	h := NewHeartbeatStore()
	if !h.IsExpired("ghost-job") {
		t.Fatal("unknown job should be considered expired")
	}
}

func TestHeartbeatStore_Beat_NotExpired(t *testing.T) {
	h := NewHeartbeatStore()
	h.Beat("nightly-backup", 5*time.Minute)
	if h.IsExpired("nightly-backup") {
		t.Fatal("freshly beaten job should not be expired")
	}
}

func TestHeartbeatStore_Beat_Expired(t *testing.T) {
	h := NewHeartbeatStore()
	h.Beat("stale-job", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	if !h.IsExpired("stale-job") {
		t.Fatal("job past TTL should be expired")
	}
}

func TestHeartbeatStore_Get_SetsExpiredFlag(t *testing.T) {
	h := NewHeartbeatStore()
	h.Beat("quick", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	rec, ok := h.Get("quick")
	if !ok {
		t.Fatal("expected record to exist")
	}
	if !rec.Expired {
		t.Fatal("expected Expired=true after TTL elapsed")
	}
}

func TestHeartbeatStore_Beat_ResetsExpiry(t *testing.T) {
	h := NewHeartbeatStore()
	h.Beat("resettable", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	// Re-beat with generous TTL — should no longer be expired.
	h.Beat("resettable", 5*time.Minute)
	if h.IsExpired("resettable") {
		t.Fatal("re-beaten job should not be expired")
	}
}

func TestHeartbeatStore_All_ReturnsSnapshot(t *testing.T) {
	h := NewHeartbeatStore()
	h.Beat("job-a", time.Minute)
	h.Beat("job-b", time.Minute)
	all := h.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 records, got %d", len(all))
	}
}
