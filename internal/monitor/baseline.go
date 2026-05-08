package monitor

import (
	"sync"
	"time"
)

// BaselineEntry holds the computed baseline (expected duration) for a job.
type BaselineEntry struct {
	JobName       string
	AvgDuration   time.Duration
	SampleCount   int
	LastUpdated   time.Time
}

// BaselineStore tracks rolling average run durations per job.
type BaselineStore struct {
	mu      sync.RWMutex
	entries map[string]*BaselineEntry
	maxSamples int
}

// NewBaselineStore creates a BaselineStore with the given rolling window size.
func NewBaselineStore(maxSamples int) *BaselineStore {
	if maxSamples <= 0 {
		maxSamples = 10
	}
	return &BaselineStore{
		entries:    make(map[string]*BaselineEntry),
		maxSamples: maxSamples,
	}
}

// Record updates the rolling average for a job given a new observed duration.
func (s *BaselineStore) Record(job string, d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.entries[job]
	if !ok {
		e = &BaselineEntry{JobName: job}
		s.entries[job] = e
	}

	if e.SampleCount < s.maxSamples {
		e.SampleCount++
	}
	total := e.AvgDuration*time.Duration(e.SampleCount-1) + d
	e.AvgDuration = total / time.Duration(e.SampleCount)
	e.LastUpdated = time.Now()
}

// Get returns the BaselineEntry for a job, and whether it exists.
func (s *BaselineStore) Get(job string) (BaselineEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	if !ok {
		return BaselineEntry{}, false
	}
	return *e, true
}

// All returns a snapshot of all baseline entries.
func (s *BaselineStore) All() []BaselineEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]BaselineEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, *e)
	}
	return out
}

// Reset clears the baseline for a specific job.
func (s *BaselineStore) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}
