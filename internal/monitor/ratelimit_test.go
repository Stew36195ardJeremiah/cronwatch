package monitor

import (
	"testing"
	"time"
)

func TestNewRateLimitStore_Defaults(t *testing.T) {
	r := NewRateLimitStore(0, 0)
	if r.window != 10*time.Minute {
		t.Errorf("expected default window 10m, got %v", r.window)
	}
	if r.maxCount != 3 {
		t.Errorf("expected default maxCount 3, got %d", r.maxCount)
	}
}

func TestRateLimitStore_FirstAlertAlwaysAllowed(t *testing.T) {
	r := NewRateLimitStore(time.Minute, 3)
	if !r.Allow("job-a") {
		t.Error("expected first alert to be allowed")
	}
}

func TestRateLimitStore_ThrottlesAfterMax(t *testing.T) {
	r := NewRateLimitStore(time.Minute, 2)
	if !r.Allow("job-b") {
		t.Error("expected 1st alert allowed")
	}
	if !r.Allow("job-b") {
		t.Error("expected 2nd alert allowed")
	}
	if r.Allow("job-b") {
		t.Error("expected 3rd alert to be throttled")
	}
}

func TestRateLimitStore_WindowExpiry(t *testing.T) {
	r := NewRateLimitStore(50*time.Millisecond, 1)
	r.Allow("job-c")
	if r.Allow("job-c") {
		t.Error("expected second alert within window to be throttled")
	}
	time.Sleep(60 * time.Millisecond)
	if !r.Allow("job-c") {
		t.Error("expected alert allowed after window expiry")
	}
}

func TestRateLimitStore_Reset_ClearsState(t *testing.T) {
	r := NewRateLimitStore(time.Minute, 1)
	r.Allow("job-d")
	if r.Allow("job-d") {
		t.Error("expected throttle before reset")
	}
	r.Reset("job-d")
	if !r.Allow("job-d") {
		t.Error("expected allow after reset")
	}
}

func TestRateLimitStore_All_ReturnsSnapshot(t *testing.T) {
	r := NewRateLimitStore(time.Minute, 5)
	r.Allow("job-e")
	r.Allow("job-f")
	all := r.All()
	if len(all) != 2 {
		t.Errorf("expected 2 entries, got %d", len(all))
	}
	if all["job-e"].Count != 1 {
		t.Errorf("expected count 1 for job-e, got %d", all["job-e"].Count)
	}
}

func TestRateLimitStore_IndependentJobs(t *testing.T) {
	r := NewRateLimitStore(time.Minute, 1)
	r.Allow("job-x")
	if !r.Allow("job-y") {
		t.Error("expected job-y to be independent of job-x throttle")
	}
}
