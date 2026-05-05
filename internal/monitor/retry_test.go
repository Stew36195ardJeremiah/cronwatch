package monitor

import (
	"testing"
	"time"
)

func TestNewRetryStore_Defaults(t *testing.T) {
	r := NewRetryStore(RetryPolicy{})
	if r.policy.MaxAttempts != 3 {
		t.Errorf("expected default MaxAttempts=3, got %d", r.policy.MaxAttempts)
	}
	if r.policy.Interval != 5*time.Minute {
		t.Errorf("expected default Interval=5m, got %v", r.policy.Interval)
	}
}

func TestRetryStore_FirstAttemptAlwaysTrue(t *testing.T) {
	r := NewRetryStore(RetryPolicy{MaxAttempts: 3, Interval: time.Minute})
	if !r.ShouldRetry("job-a") {
		t.Error("expected first attempt to return true")
	}
	state, ok := r.State("job-a")
	if !ok {
		t.Fatal("expected state to exist after first attempt")
	}
	if state.Attempts != 1 {
		t.Errorf("expected Attempts=1, got %d", state.Attempts)
	}
}

func TestRetryStore_IntervalThrottles(t *testing.T) {
	r := NewRetryStore(RetryPolicy{MaxAttempts: 5, Interval: time.Hour})
	r.ShouldRetry("job-b") // first attempt
	if r.ShouldRetry("job-b") {
		t.Error("expected second attempt within interval to return false")
	}
}

func TestRetryStore_ExhaustsAfterMaxAttempts(t *testing.T) {
	r := NewRetryStore(RetryPolicy{MaxAttempts: 2, Interval: 0})
	r.ShouldRetry("job-c") // attempt 1
	r.ShouldRetry("job-c") // attempt 2 — exhausted
	if r.ShouldRetry("job-c") {
		t.Error("expected exhausted job to return false")
	}
	state, _ := r.State("job-c")
	if !state.Exhausted {
		t.Error("expected Exhausted=true")
	}
}

func TestRetryStore_Reset_ClearsState(t *testing.T) {
	r := NewRetryStore(RetryPolicy{MaxAttempts: 2, Interval: 0})
	r.ShouldRetry("job-d")
	r.ShouldRetry("job-d")
	r.Reset("job-d")
	_, ok := r.State("job-d")
	if ok {
		t.Error("expected state to be cleared after Reset")
	}
	// After reset, should be retryable again
	if !r.ShouldRetry("job-d") {
		t.Error("expected ShouldRetry=true after Reset")
	}
}

func TestRetryStore_StateUnknownJob(t *testing.T) {
	r := NewRetryStore(RetryPolicy{MaxAttempts: 3, Interval: time.Minute})
	_, ok := r.State("nonexistent")
	if ok {
		t.Error("expected ok=false for unknown job")
	}
}
