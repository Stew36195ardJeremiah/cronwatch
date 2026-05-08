package monitor

import (
	"testing"
	"time"
)

func TestNewQuotaStore_Defaults(t *testing.T) {
	q := NewQuotaStore(5, time.Minute)
	if q.defaultLimit != 5 {
		t.Errorf("expected defaultLimit 5, got %d", q.defaultLimit)
	}
	if q.defaultWindow != time.Minute {
		t.Errorf("expected defaultWindow 1m, got %v", q.defaultWindow)
	}
	if len(q.entries) != 0 {
		t.Error("expected empty entries")
	}
}

func TestQuotaStore_FirstAlertAllowed(t *testing.T) {
	q := NewQuotaStore(3, time.Minute)
	if !q.Allow("job-a") {
		t.Error("expected first alert to be allowed")
	}
}

func TestQuotaStore_ExhaustsQuota(t *testing.T) {
	q := NewQuotaStore(2, time.Minute)
	if !q.Allow("job-b") {
		t.Error("expected allow on attempt 1")
	}
	if !q.Allow("job-b") {
		t.Error("expected allow on attempt 2")
	}
	if q.Allow("job-b") {
		t.Error("expected deny on attempt 3 (over quota)")
	}
}

func TestQuotaStore_WindowExpiry(t *testing.T) {
	q := NewQuotaStore(1, 10*time.Millisecond)
	if !q.Allow("job-c") {
		t.Error("expected allow on first attempt")
	}
	if q.Allow("job-c") {
		t.Error("expected deny after quota exhausted")
	}
	time.Sleep(20 * time.Millisecond)
	if !q.Allow("job-c") {
		t.Error("expected allow after window reset")
	}
}

func TestQuotaStore_SetLimit_Override(t *testing.T) {
	q := NewQuotaStore(1, time.Minute)
	q.SetLimit("job-d", 3, time.Minute)
	for i := 0; i < 3; i++ {
		if !q.Allow("job-d") {
			t.Errorf("expected allow on attempt %d", i+1)
		}
	}
	if q.Allow("job-d") {
		t.Error("expected deny after custom limit")
	}
}

func TestQuotaStore_Get_UnknownJob(t *testing.T) {
	q := NewQuotaStore(5, time.Minute)
	if e := q.Get("unknown"); e != nil {
		t.Errorf("expected nil for unknown job, got %+v", e)
	}
}

func TestQuotaStore_Get_KnownJob(t *testing.T) {
	q := NewQuotaStore(5, time.Minute)
	q.Allow("job-e")
	e := q.Get("job-e")
	if e == nil {
		t.Fatal("expected entry for job-e")
	}
	if e.Count != 1 {
		t.Errorf("expected count 1, got %d", e.Count)
	}
}

func TestQuotaStore_Reset_ClearsCount(t *testing.T) {
	q := NewQuotaStore(1, time.Minute)
	q.Allow("job-f")
	if q.Allow("job-f") {
		t.Error("expected deny before reset")
	}
	q.Reset("job-f")
	if !q.Allow("job-f") {
		t.Error("expected allow after reset")
	}
}

func TestQuotaStore_All_ReturnsSnapshot(t *testing.T) {
	q := NewQuotaStore(5, time.Minute)
	q.Allow("job-g")
	q.Allow("job-h")
	all := q.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
}
