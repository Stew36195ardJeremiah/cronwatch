package monitor

import (
	"testing"
	"time"
)

func TestNewEscalationStore_Empty(t *testing.T) {
	es := NewEscalationStore()
	if len(es.All()) != 0 {
		t.Fatal("expected empty store")
	}
}

func TestEscalationLevel_NoPolicyReturnsNone(t *testing.T) {
	es := NewEscalationStore()
	es.Trigger("job-a")
	if lvl := es.Level("job-a"); lvl != EscalationNone {
		t.Fatalf("expected None without policy, got %s", lvl)
	}
}

func TestEscalationLevel_NoTriggerReturnsNone(t *testing.T) {
	es := NewEscalationStore()
	es.SetPolicy(EscalationPolicy{
		JobName:   "job-a",
		WarnAfter: time.Millisecond,
	})
	if lvl := es.Level("job-a"); lvl != EscalationNone {
		t.Fatalf("expected None without trigger, got %s", lvl)
	}
}

func TestEscalationLevel_WarnAfterElapsed(t *testing.T) {
	es := NewEscalationStore()
	es.SetPolicy(EscalationPolicy{
		JobName:       "job-b",
		WarnAfter:     time.Millisecond,
		CriticalAfter: time.Hour,
	})
	es.Trigger("job-b")
	time.Sleep(5 * time.Millisecond)
	if lvl := es.Level("job-b"); lvl != EscalationWarn {
		t.Fatalf("expected Warn, got %s", lvl)
	}
}

func TestEscalationLevel_CriticalAfterElapsed(t *testing.T) {
	es := NewEscalationStore()
	es.SetPolicy(EscalationPolicy{
		JobName:       "job-c",
		WarnAfter:     time.Millisecond,
		CriticalAfter: 2 * time.Millisecond,
	})
	es.Trigger("job-c")
	time.Sleep(10 * time.Millisecond)
	if lvl := es.Level("job-c"); lvl != EscalationCritical {
		t.Fatalf("expected Critical, got %s", lvl)
	}
}

func TestEscalationReset_ClearsState(t *testing.T) {
	es := NewEscalationStore()
	es.SetPolicy(EscalationPolicy{
		JobName:   "job-d",
		WarnAfter: time.Millisecond,
	})
	es.Trigger("job-d")
	time.Sleep(5 * time.Millisecond)
	es.Reset("job-d")
	if lvl := es.Level("job-d"); lvl != EscalationNone {
		t.Fatalf("expected None after reset, got %s", lvl)
	}
	if len(es.All()) != 0 {
		t.Fatal("expected All() empty after reset")
	}
}

func TestEscalationTrigger_IdempotentTimestamp(t *testing.T) {
	es := NewEscalationStore()
	es.SetPolicy(EscalationPolicy{
		JobName:       "job-e",
		WarnAfter:     time.Millisecond,
		CriticalAfter: time.Hour,
	})
	es.Trigger("job-e")
	first := es.triggered["job-e"]
	time.Sleep(2 * time.Millisecond)
	es.Trigger("job-e") // should not overwrite
	if !es.triggered["job-e"].Equal(first) {
		t.Fatal("second Trigger should not update timestamp")
	}
}

func TestEscalationLevelString(t *testing.T) {
	if EscalationNone.String() != "none" {
		t.Errorf("unexpected: %s", EscalationNone.String())
	}
	if EscalationWarn.String() != "warn" {
		t.Errorf("unexpected: %s", EscalationWarn.String())
	}
	if EscalationCritical.String() != "critical" {
		t.Errorf("unexpected: %s", EscalationCritical.String())
	}
}
