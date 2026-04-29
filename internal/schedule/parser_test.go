package schedule_test

import (
	"testing"
	"time"

	"github.com/cronwatch/cronwatch/internal/schedule"
)

func TestNextRun_Valid(t *testing.T) {
	// Every minute: next run should be within the next 60 seconds
	from := time.Now().Truncate(time.Minute)
	next, err := schedule.NextRun("* * * * *", from)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !next.After(from) {
		t.Errorf("expected next run to be after %v, got %v", from, next)
	}
	if next.Sub(from) > 2*time.Minute {
		t.Errorf("expected next run within 2 minutes, got %v", next.Sub(from))
	}
}

func TestNextRun_InvalidExpr(t *testing.T) {
	_, err := schedule.NextRun("not-a-cron", time.Now())
	if err == nil {
		t.Fatal("expected error for invalid cron expression, got nil")
	}
}

func TestIsOverdue_NotOverdue(t *testing.T) {
	// Job ran just now; with a 5-minute tolerance it should not be overdue
	lastRun := time.Now().Add(-30 * time.Second)
	overdue, err := schedule.IsOverdue("* * * * *", lastRun, 5*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if overdue {
		t.Error("expected job to not be overdue")
	}
}

func TestIsOverdue_Overdue(t *testing.T) {
	// Job last ran 10 minutes ago on a every-minute schedule with 0 tolerance
	lastRun := time.Now().Add(-10 * time.Minute)
	overdue, err := schedule.IsOverdue("* * * * *", lastRun, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !overdue {
		t.Error("expected job to be overdue")
	}
}

func TestDriftDuration_Positive(t *testing.T) {
	// Job ran 10 minutes ago on every-minute schedule → drift should be positive
	lastRun := time.Now().Add(-10 * time.Minute)
	drift, err := schedule.DriftDuration("* * * * *", lastRun)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if drift <= 0 {
		t.Errorf("expected positive drift, got %v", drift)
	}
}

func TestDriftDuration_InvalidExpr(t *testing.T) {
	_, err := schedule.DriftDuration("bad expr", time.Now())
	if err == nil {
		t.Fatal("expected error for invalid cron expression")
	}
}
