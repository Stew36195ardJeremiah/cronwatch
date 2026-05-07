package monitor

import (
	"testing"
	"time"
)

func TestNewCooldownStore_DefaultDuration(t *testing.T) {
	cs := NewCooldownStore(0)
	if cs.defaultDuration != 5*time.Minute {
		t.Errorf("expected default 5m, got %v", cs.defaultDuration)
	}
}

func TestCooldownStore_NotInCooldownInitially(t *testing.T) {
	cs := NewCooldownStore(time.Minute)
	if cs.InCooldown("backup") {
		t.Error("expected no cooldown for unknown job")
	}
}

func TestCooldownStore_ActivateAndCheck(t *testing.T) {
	cs := NewCooldownStore(time.Minute)
	cs.Activate("backup")
	if !cs.InCooldown("backup") {
		t.Error("expected job to be in cooldown after Activate")
	}
}

func TestCooldownStore_ActivateFor_CustomDuration(t *testing.T) {
	cs := NewCooldownStore(time.Minute)
	cs.ActivateFor("sync", 10*time.Millisecond)
	if !cs.InCooldown("sync") {
		t.Error("expected cooldown to be active immediately")
	}
	time.Sleep(20 * time.Millisecond)
	if cs.InCooldown("sync") {
		t.Error("expected cooldown to have expired")
	}
}

func TestCooldownStore_Lift_RemovesCooldown(t *testing.T) {
	cs := NewCooldownStore(time.Minute)
	cs.Activate("cleanup")
	cs.Lift("cleanup")
	if cs.InCooldown("cleanup") {
		t.Error("expected cooldown to be lifted")
	}
}

func TestCooldownStore_Lift_UnknownJob_NoOp(t *testing.T) {
	cs := NewCooldownStore(time.Minute)
	cs.Lift("nonexistent") // should not panic
}

func TestCooldownStore_All_ReturnsActiveOnly(t *testing.T) {
	cs := NewCooldownStore(time.Minute)
	cs.Activate("job-a")
	cs.ActivateFor("job-b", time.Millisecond)
	time.Sleep(5 * time.Millisecond)

	all := cs.All()
	if _, ok := all["job-a"]; !ok {
		t.Error("expected job-a in All()")
	}
	if _, ok := all["job-b"]; ok {
		t.Error("expected expired job-b to be absent from All()")
	}
}

func TestCooldownStore_All_RemainingDurationPositive(t *testing.T) {
	cs := NewCooldownStore(time.Minute)
	cs.Activate("report")
	all := cs.All()
	if all["report"] <= 0 {
		t.Errorf("expected positive remaining duration, got %v", all["report"])
	}
}
