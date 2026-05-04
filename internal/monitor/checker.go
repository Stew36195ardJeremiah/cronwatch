package monitor

import (
	"fmt"
	"log"
	"time"

	"github.com/yourusername/cronwatch/internal/alert"
	"github.com/yourusername/cronwatch/internal/schedule"
)

// CheckResult holds the outcome of a single job check.
type CheckResult struct {
	JobName   string
	Overdue   bool
	Drift     time.Duration
	LastRun   time.Time
	NextRun   time.Time
	CheckedAt time.Time
}

// String returns a human-readable summary of the check result.
func (r CheckResult) String() string {
	if r.Overdue {
		return fmt.Sprintf(
			"job %q is OVERDUE (drift: %s, last run: %s, expected next: %s)",
			r.JobName, r.Drift.Round(time.Second), r.LastRun.Format(time.RFC3339), r.NextRun.Format(time.RFC3339),
		)
	}
	return fmt.Sprintf(
		"job %q is OK (next run: %s)",
		r.JobName, r.NextRun.Format(time.RFC3339),
	)
}

// CheckJobs evaluates all tracked jobs and dispatches alerts for any that are
// overdue or have drifted beyond the configured threshold. It returns the list
// of results for all jobs that were checked.
func (m *Monitor) CheckJobs() []CheckResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	results := make([]CheckResult, 0, len(m.cfg.Jobs))
	now := time.Now()

	for _, job := range m.cfg.Jobs {
		status, ok := m.statuses[job.Name]
		if !ok {
			log.Printf("[checker] no status found for job %q, skipping", job.Name)
			continue
		}

		nextRun, err := schedule.NextRun(job.Schedule, status.LastRun)
		if err != nil {
			m.alerts.Error(fmt.Sprintf(
				"job %q has an invalid schedule expression %q: %v",
				job.Name, job.Schedule, err,
			))
			continue
		}

		overdue := schedule.IsOverdue(job.Schedule, status.LastRun, now)
		drift := schedule.DriftDuration(nextRun, now)

		result := CheckResult{
			JobName:   job.Name,
			Overdue:   overdue,
			Drift:     drift,
			LastRun:   status.LastRun,
			NextRun:   nextRun,
			CheckedAt: now,
		}
		results = append(results, result)

		if overdue {
			m.dispatchOverdueAlert(job.Name, result)
		} else if job.DriftThreshold > 0 && drift > job.DriftThreshold {
			m.dispatchDriftAlert(job.Name, result, job.DriftThreshold)
		}
	}

	return results
}

// dispatchOverdueAlert sends an error-level alert for a job that has not run
// within its expected schedule window.
func (m *Monitor) dispatchOverdueAlert(jobName string, r CheckResult) {
	msg := fmt.Sprintf(
		"OVERDUE: job %q has not run since %s (drift: %s)",
		jobName,
		r.LastRun.Format(time.RFC3339),
		r.Drift.Round(time.Second),
	)
	log.Printf("[checker] %s", msg)
	m.alerts.Error(msg)
}

// dispatchDriftAlert sends a warning-level alert when a job's execution drift
// exceeds the configured threshold but the job has not yet been marked overdue.
func (m *Monitor) dispatchDriftAlert(jobName string, r CheckResult, threshold time.Duration) {
	msg := fmt.Sprintf(
		"DRIFT WARNING: job %q drift is %s (threshold: %s, next run: %s)",
		jobName,
		r.Drift.Round(time.Second),
		threshold.Round(time.Second),
		r.NextRun.Format(time.RFC3339),
	)
	log.Printf("[checker] %s", msg)
	m.alerts.Warn(msg)
}

// alertDispatcher is a small interface so checker logic can be tested without
// a full alert.Manager.
type alertDispatcher interface {
	Warn(msg string)
	Error(msg string)
}

// Ensure *alert.Manager satisfies the interface at compile time.
var _ alertDispatcher = (*alert.Manager)(nil)
