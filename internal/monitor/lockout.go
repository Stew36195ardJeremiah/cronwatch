package monitor

import (
	"sync"
	"time"
)

// LockoutEntry represents a job that has been locked out from alerting
// after too many consecutive failures.
type LockoutEntry struct {
	Job       string
	LockedAt  time.Time
	Until     time.Time
	Reason    string
	Trips     int
}

// LockoutStore tracks jobs that are temporarily locked out from alerting
// due to repeated failures exceeding a configurable threshold.
type LockoutStore struct {
	mu        sync.RWMutex
	entries   map[string]*LockoutEntry
	threshold int
	duration  time.Duration
	failures  map[string]int
}

// NewLockoutStore creates a new LockoutStore with the given failure threshold
// and lockout duration.
func NewLockoutStore(threshold int, duration time.Duration) *LockoutStore {
	if threshold <= 0 {
		threshold = 5
	}
	if duration <= 0 {
		duration = 15 * time.Minute
	}
	return &LockoutStore{
		entries:   make(map[string]*LockoutEntry),
		threshold: threshold,
		duration:  duration,
		failures:  make(map[string]int),
	}
}

// RecordFailure increments the failure counter for a job. If the threshold is
// reached, the job is locked out for the configured duration.
func (s *LockoutStore) RecordFailure(job, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures[job]++
	if s.failures[job] >= s.threshold {
		now := time.Now()
		trips := 0
		if existing, ok := s.entries[job]; ok {
			trips = existing.Trips
		}
		s.entries[job] = &LockoutEntry{
			Job:      job,
			LockedAt: now,
			Until:    now.Add(s.duration),
			Reason:   reason,
			Trips:    trips + 1,
		}
		s.failures[job] = 0
	}
}

// IsLockedOut returns true if the job is currently locked out.
func (s *LockoutStore) IsLockedOut(job string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entry, ok := s.entries[job]
	if !ok {
		return false
	}
	return time.Now().Before(entry.Until)
}

// Get returns the lockout entry for a job, or nil if not found.
func (s *LockoutStore) Get(job string) *LockoutEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	if !ok {
		return nil
	}
	copy := *e
	return &copy
}

// Lift removes any active lockout for the given job and resets its failure count.
func (s *LockoutStore) Lift(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
	s.failures[job] = 0
}

// All returns a snapshot of all lockout entries.
func (s *LockoutStore) All() []LockoutEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]LockoutEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, *e)
	}
	return out
}
