package monitor

import (
	"testing"
	"time"
)

func newSnapshotStores() (*StatusStore, *MetricsStore, *SuppressionStore) {
	return NewStatusStore(), NewMetricsStore(), NewSuppressionStore()
}

func TestSnapshot_EmptyStores(t *testing.T) {
	ss, ms, sup := newSnapshotStores()
	b := NewSnapshotBuilder(ss, ms, sup)
	snap := b.Build()

	if snap.CapturedAt.IsZero() {
		t.Error("expected CapturedAt to be set")
	}
	if len(snap.Statuses) != 0 {
		t.Errorf("expected 0 statuses, got %d", len(snap.Statuses))
	}
	if len(snap.Metrics) != 0 {
		t.Errorf("expected 0 metrics, got %d", len(snap.Metrics))
	}
	if len(snap.Suppressed) != 0 {
		t.Errorf("expected 0 suppressed, got %d", len(snap.Suppressed))
	}
}

func TestSnapshot_CapturesStatuses(t *testing.T) {
	ss, ms, sup := newSnapshotStores()
	ss.Set("job-a", JobStatus{Name: "job-a", Healthy: true})
	ss.Set("job-b", JobStatus{Name: "job-b", Healthy: false})

	b := NewSnapshotBuilder(ss, ms, sup)
	snap := b.Build()

	if len(snap.Statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(snap.Statuses))
	}
	if !snap.Statuses["job-a"].Healthy {
		t.Error("expected job-a to be healthy")
	}
}

func TestSnapshot_CapturesMetrics(t *testing.T) {
	ss, ms, sup := newSnapshotStores()
	ss.Set("job-a", JobStatus{Name: "job-a", Healthy: true})
	ms.Record("job-a", 10*time.Second, false)
	ms.Record("job-a", 20*time.Second, false)

	b := NewSnapshotBuilder(ss, ms, sup)
	snap := b.Build()

	m, ok := snap.Metrics["job-a"]
	if !ok {
		t.Fatal("expected metrics for job-a")
	}
	if m.RunCount != 2 {
		t.Errorf("expected RunCount=2, got %d", m.RunCount)
	}
}

func TestSnapshot_CapturesSuppression(t *testing.T) {
	ss, ms, sup := newSnapshotStores()
	ss.Set("job-a", JobStatus{Name: "job-a", Healthy: true})
	sup.Suppress("job-a", 5*time.Minute)

	b := NewSnapshotBuilder(ss, ms, sup)
	snap := b.Build()

	if _, ok := snap.Suppressed["job-a"]; !ok {
		t.Error("expected job-a to appear in suppressed map")
	}
}

func TestSnapshot_MetricsOnlyForKnownJobs(t *testing.T) {
	ss, ms, sup := newSnapshotStores()
	// Record metrics for a job not in status store
	ms.Record("ghost-job", 5*time.Second, false)

	b := NewSnapshotBuilder(ss, ms, sup)
	snap := b.Build()

	if _, ok := snap.Metrics["ghost-job"]; ok {
		t.Error("ghost-job should not appear in snapshot metrics")
	}
}
