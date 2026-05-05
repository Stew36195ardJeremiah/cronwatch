package monitor

import (
	"testing"
	"time"
)

func TestNewSuppressionStore_Empty(t *testing.T) {
	s := NewSuppressionStore()
	if s.IsSuppressed("any-job") {
		t.Fatal("expected no suppression on empty store")
	}
}

func TestSuppress_BlocksAlerts(t *testing.T) {
	s := NewSuppressionStore()
	s.Suppress("backup", 5*time.Minute)
	if !s.IsSuppressed("backup") {
		t.Fatal("expected job to be suppressed")
	}
}

func TestSuppress_Expires(t *testing.T) {
	s := NewSuppressionStore()
	s.Suppress("backup", -1*time.Millisecond) // already expired
	if s.IsSuppressed("backup") {
		t.Fatal("expected suppression to have expired")
	}
}

func TestLift_RemovesSuppression(t *testing.T) {
	s := NewSuppressionStore()
	s.Suppress("backup", 10*time.Minute)
	s.Lift("backup")
	if s.IsSuppressed("backup") {
		t.Fatal("expected suppression to be lifted")
	}
}

func TestAll_ReturnsSnapshot(t *testing.T) {
	s := NewSuppressionStore()
	s.Suppress("job-a", time.Minute)
	s.Suppress("job-b", time.Minute)
	all := s.All()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if _, ok := all["job-a"]; !ok {
		t.Error("expected job-a in snapshot")
	}
	if _, ok := all["job-b"]; !ok {
		t.Error("expected job-b in snapshot")
	}
}

func TestAll_DoesNotLeakReference(t *testing.T) {
	s := NewSuppressionStore()
	s.Suppress("job-a", time.Minute)
	all := s.All()
	delete(all, "job-a")
	if !s.IsSuppressed("job-a") {
		t.Fatal("modifying snapshot should not affect store")
	}
}
