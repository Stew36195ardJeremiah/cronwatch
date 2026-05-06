package monitor

import (
	"fmt"
	"sync"
	"time"
)

// DependencyEdge represents a directed dependency between two jobs.
// Job B depends on Job A means B should not run until A has succeeded.
type DependencyEdge struct {
	From string // upstream job (must complete first)
	To   string // downstream job (depends on From)
}

// DependencyStore tracks inter-job dependencies and the last successful
// completion time of each job so downstream jobs can be gated.
type DependencyStore struct {
	mu       sync.RWMutex
	edges    []DependencyEdge
	lastOK   map[string]time.Time
}

// NewDependencyStore creates an empty DependencyStore.
func NewDependencyStore() *DependencyStore {
	return &DependencyStore{
		edges:  []DependencyEdge{},
		lastOK: make(map[string]time.Time),
	}
}

// AddEdge registers that job `to` depends on job `from`.
func (d *DependencyStore) AddEdge(from, to string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.edges = append(d.edges, DependencyEdge{From: from, To: to})
}

// MarkSuccess records that a job completed successfully at the given time.
func (d *DependencyStore) MarkSuccess(job string, at time.Time) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.lastOK[job] = at
}

// Blocked returns an error if any upstream dependency of `job` has not
// completed successfully since `since`. Returns nil when the job is clear to run.
func (d *DependencyStore) Blocked(job string, since time.Time) error {
	d.mu.RLock()
	defer d.mu.RUnlock()
	for _, e := range d.edges {
		if e.To != job {
			continue
		}
		t, ok := d.lastOK[e.From]
		if !ok || t.Before(since) {
			return fmt.Errorf("job %q is blocked: upstream %q has not succeeded since %s", job, e.From, since.Format(time.RFC3339))
		}
	}
	return nil
}

// Edges returns a snapshot of all registered dependency edges.
func (d *DependencyStore) Edges() []DependencyEdge {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]DependencyEdge, len(d.edges))
	copy(out, d.edges)
	return out
}
