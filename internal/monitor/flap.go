package monitor

import (
	"sync"
	"time"
)

// FlapEntry tracks state transitions for a job to detect flapping.
type FlapEntry struct {
	Transitions []time.Time
	IsFlapping  bool
	LastChecked time.Time
}

// FlapStore detects jobs that oscillate between healthy and failing states.
type FlapStore struct {
	mu       sync.RWMutex
	entries  map[string]*FlapEntry
	window   time.Duration
	threshold int
}

// NewFlapStore creates a FlapStore with the given detection window and
// minimum number of transitions required to declare flapping.
func NewFlapStore(window time.Duration, threshold int) *FlapStore {
	if threshold <= 0 {
		threshold = 4
	}
	if window <= 0 {
		window = 10 * time.Minute
	}
	return &FlapStore{
		entries:   make(map[string]*FlapEntry),
		window:    window,
		threshold: threshold,
	}
}

// RecordTransition records a state change for the given job and returns
// true if the job is currently considered to be flapping.
func (f *FlapStore) RecordTransition(job string, at time.Time) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	e, ok := f.entries[job]
	if !ok {
		e = &FlapEntry{}
		f.entries[job] = e
	}

	// Append and prune old transitions outside the window.
	e.Transitions = append(e.Transitions, at)
	cutoff := at.Add(-f.window)
	filtered := e.Transitions[:0]
	for _, t := range e.Transitions {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	e.Transitions = filtered
	e.IsFlapping = len(e.Transitions) >= f.threshold
	e.LastChecked = at
	return e.IsFlapping
}

// IsFlapping returns whether the job is currently in a flapping state.
func (f *FlapStore) IsFlapping(job string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	e, ok := f.entries[job]
	if !ok {
		return false
	}
	return e.IsFlapping
}

// Get returns the FlapEntry for the given job, or nil if unknown.
func (f *FlapStore) Get(job string) *FlapEntry {
	f.mu.RLock()
	defer f.mu.RUnlock()
	e, ok := f.entries[job]
	if !ok {
		return nil
	}
	copy := *e
	return &copy
}

// Reset clears flap state for a job.
func (f *FlapStore) Reset(job string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.entries, job)
}

// All returns a snapshot of all flap entries.
func (f *FlapStore) All() map[string]FlapEntry {
	f.mu.RLock()
	defer f.mu.RUnlock()
	out := make(map[string]FlapEntry, len(f.entries))
	for k, v := range f.entries {
		out[k] = *v
	}
	return out
}
