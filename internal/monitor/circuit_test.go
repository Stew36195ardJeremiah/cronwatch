package monitor

import (
	"testing"
	"time"
)

func TestNewCircuitBreakerStore_Defaults(t *testing.T) {
	cb := NewCircuitBreakerStore(0, 0)
	if cb.maxFailures != 3 {
		t.Errorf("expected default maxFailures=3, got %d", cb.maxFailures)
	}
	if cb.recoverIn != 5*time.Minute {
		t.Errorf("expected default recoverIn=5m, got %v", cb.recoverIn)
	}
}

func TestCircuitBreaker_InitiallyClosed(t *testing.T) {
	cb := NewCircuitBreakerStore(3, time.Minute)
	if s := cb.State("job1"); s != CircuitClosed {
		t.Errorf("expected closed, got %s", s)
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreakerStore(3, time.Minute)
	cb.RecordFailure("job1")
	cb.RecordFailure("job1")
	if cb.State("job1") != CircuitClosed {
		t.Error("expected still closed after 2 failures")
	}
	cb.RecordFailure("job1")
	if cb.State("job1") != CircuitOpen {
		t.Error("expected open after 3 failures")
	}
}

func TestCircuitBreaker_SuccessClosesClosed(t *testing.T) {
	cb := NewCircuitBreakerStore(3, time.Minute)
	cb.RecordFailure("job1")
	cb.RecordSuccess("job1")
	if cb.State("job1") != CircuitClosed {
		t.Error("expected closed after success")
	}
}

func TestCircuitBreaker_TransitionsToHalfOpen(t *testing.T) {
	cb := NewCircuitBreakerStore(1, 10*time.Millisecond)
	cb.RecordFailure("job1")
	if cb.State("job1") != CircuitOpen {
		t.Fatal("expected open")
	}
	time.Sleep(20 * time.Millisecond)
	if cb.State("job1") != CircuitHalfOpen {
		t.Error("expected half-open after recovery window")
	}
}

func TestCircuitBreaker_HalfOpenClosedOnSuccess(t *testing.T) {
	cb := NewCircuitBreakerStore(1, 10*time.Millisecond)
	cb.RecordFailure("job1")
	time.Sleep(20 * time.Millisecond)
	_ = cb.State("job1") // trigger half-open
	cb.RecordSuccess("job1")
	if cb.State("job1") != CircuitClosed {
		t.Error("expected closed after success in half-open")
	}
}

func TestCircuitBreaker_Reset_ClearsState(t *testing.T) {
	cb := NewCircuitBreakerStore(1, time.Minute)
	cb.RecordFailure("job1")
	cb.Reset("job1")
	if cb.State("job1") != CircuitClosed {
		t.Error("expected closed after reset")
	}
}

func TestCircuitBreaker_All_ReturnsSnapshot(t *testing.T) {
	cb := NewCircuitBreakerStore(1, time.Minute)
	cb.RecordFailure("jobA")
	cb.RecordFailure("jobB")
	all := cb.All()
	if _, ok := all["jobA"]; !ok {
		t.Error("expected jobA in snapshot")
	}
	if _, ok := all["jobB"]; !ok {
		t.Error("expected jobB in snapshot")
	}
}

func TestCircuitState_String(t *testing.T) {
	cases := map[CircuitState]string{
		CircuitClosed:   "closed",
		CircuitOpen:     "open",
		CircuitHalfOpen: "half-open",
		CircuitState(99): "unknown",
	}
	for state, want := range cases {
		if got := state.String(); got != want {
			t.Errorf("state %d: got %q, want %q", state, got, want)
		}
	}
}
