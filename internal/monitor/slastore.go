package monitor

import (
	"sync"
	"time"
)

// SLAEntry holds the SLA configuration and current compliance state for a job.
type SLAEntry struct {
	Job           string
	MaxDuration   time.Duration
	Deadline      time.Time // wall-clock deadline within a period (zero = unset)
	ViolationCount int
	LastViolation  time.Time
	Compliant      bool
}

// SLAStore tracks per-job SLA policies and violation history.
type SLAStore struct {
	mu      sync.RWMutex
	entries map[string]*SLAEntry
}

// NewSLAStore returns an initialised SLAStore.
func NewSLAStore() *SLAStore {
	return &SLAStore{entries: make(map[string]*SLAEntry)}
}

// Set registers or updates the SLA policy for a job.
func (s *SLAStore) Set(job string, maxDuration time.Duration, deadline time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		e = &SLAEntry{Job: job, Compliant: true}
		s.entries[job] = e
	}
	e.MaxDuration = maxDuration
	e.Deadline = deadline
}

// RecordRun evaluates whether the completed run violates the SLA and updates state.
func (s *SLAStore) RecordRun(job string, duration time.Duration, finishedAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.entries[job]
	if !ok {
		return
	}
	violated := false
	if e.MaxDuration > 0 && duration > e.MaxDuration {
		violated = true
	}
	if !e.Deadline.IsZero() && finishedAt.After(e.Deadline) {
		violated = true
	}
	if violated {
		e.ViolationCount++
		e.LastViolation = finishedAt
		e.Compliant = false
	} else {
		e.Compliant = true
	}
}

// Get returns a copy of the SLAEntry for the given job, and whether it exists.
func (s *SLAStore) Get(job string) (SLAEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	if !ok {
		return SLAEntry{}, false
	}
	return *e, true
}

// All returns a snapshot of all SLA entries.
func (s *SLAStore) All() []SLAEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SLAEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, *e)
	}
	return out
}

// Remove deletes the SLA policy for a job.
func (s *SLAStore) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}
