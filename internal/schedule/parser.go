package schedule

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// NextRun returns the next scheduled run time after 'from' for the given cron expression.
func NextRun(expr string, from time.Time) (time.Time, error) {
	parser := cron.NewParser(
		cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow,
	)

	sched, err := parser.Parse(expr)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid cron expression %q: %w", expr, err)
	}

	return sched.Next(from), nil
}

// IsOverdue returns true if the job has not run within its expected window.
// expr is the cron schedule, lastRun is when the job last executed,
// and driftTolerance is how much lateness is acceptable.
func IsOverdue(expr string, lastRun time.Time, driftTolerance time.Duration) (bool, error) {
	next, err := NextRun(expr, lastRun)
	if err != nil {
		return false, err
	}

	deadline := next.Add(driftTolerance)
	return time.Now().After(deadline), nil
}

// DriftDuration returns how late the job is relative to its scheduled next run.
// A positive value means the job is overdue; negative means it is still on time.
func DriftDuration(expr string, lastRun time.Time) (time.Duration, error) {
	next, err := NextRun(expr, lastRun)
	if err != nil {
		return 0, err
	}

	return time.Since(next), nil
}
