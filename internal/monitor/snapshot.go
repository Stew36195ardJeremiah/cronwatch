package monitor

import (
	"sync"
	"time"
)

// Snapshot captures a point-in-time view of all job statuses,
// metrics, and suppression state for reporting or debugging.
type Snapshot struct {
	CapturedAt  time.Time                `json:"captured_at"`
	Statuses    map[string]JobStatus      `json:"statuses"`
	Metrics     map[string]MetricSummary  `json:"metrics"`
	Suppressed  map[string]time.Time      `json:"suppressed"`
}

// MetricSummary is a lightweight view of a job's recorded metrics.
type MetricSummary struct {
	RunCount    int           `json:"run_count"`
	FailCount   int           `json:"fail_count"`
	MaxDrift    time.Duration `json:"max_drift_ns"`
	AvgDuration time.Duration `json:"avg_duration_ns"`
}

// SnapshotBuilder assembles a Snapshot from the various stores.
type SnapshotBuilder struct {
	mu          sync.Mutex
	statuses    *StatusStore
	metrics     *MetricsStore
	suppression *SuppressionStore
}

// NewSnapshotBuilder creates a builder wired to the provided stores.
func NewSnapshotBuilder(s *StatusStore, m *MetricsStore, sup *SuppressionStore) *SnapshotBuilder {
	return &SnapshotBuilder{
		statuses:    s,
		metrics:     m,
		suppression: sup,
	}
}

// Build produces a consistent Snapshot. Stores are read independently
// (not under a single lock) so minor skew is possible, but each store
// is individually safe for concurrent access.
func (b *SnapshotBuilder) Build() Snapshot {
	b.mu.Lock()
	defer b.mu.Unlock()

	snap := Snapshot{
		CapturedAt: time.Now().UTC(),
		Statuses:   make(map[string]JobStatus),
		Metrics:    make(map[string]MetricSummary),
		Suppressed: make(map[string]time.Time),
	}

	for name, st := range b.statuses.All() {
		snap.Statuses[name] = st
	}

	for name, entry := range b.suppression.All() {
		snap.Suppressed[name] = entry
	}

	for name := range snap.Statuses {
		if m, ok := b.metrics.Get(name); ok {
			snap.Metrics[name] = MetricSummary{
				RunCount:    m.RunCount,
				FailCount:   m.FailCount,
				MaxDrift:    m.MaxDrift,
				AvgDuration: m.AvgDuration,
			}
		}
	}

	return snap
}
