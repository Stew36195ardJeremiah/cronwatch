package monitor

import (
	"sync"
	"time"
)

// AcknowledgementEntry records when a job alert was acknowledged and by whom.
type AcknowledgementEntry struct {
	Job         string
	AckedBy     string
	AckedAt     time.Time
	ExpiresAt   time.Time
	Note        string
}

// IsExpired returns true if the acknowledgement window has passed.
func (a AcknowledgementEntry) IsExpired(now time.Time) bool {
	return now.After(a.ExpiresAt)
}

// AcknowledgementStore tracks active acknowledgements for jobs.
type AcknowledgementStore struct {
	mu      sync.RWMutex
	entries map[string]AcknowledgementEntry
}

// NewAcknowledgementStore returns an initialised AcknowledgementStore.
func NewAcknowledgementStore() *AcknowledgementStore {
	return &AcknowledgementStore{
		entries: make(map[string]AcknowledgementEntry),
	}
}

// Acknowledge records an acknowledgement for the given job.
func (s *AcknowledgementStore) Acknowledge(job, ackedBy, note string, duration time.Duration) {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = AcknowledgementEntry{
		Job:       job,
		AckedBy:   ackedBy,
		AckedAt:   now,
		ExpiresAt: now.Add(duration),
		Note:      note,
	}
}

// IsAcknowledged returns true if the job has an active, non-expired acknowledgement.
func (s *AcknowledgementStore) IsAcknowledged(job string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	if !ok {
		return false
	}
	return !e.IsExpired(time.Now())
}

// Get returns the acknowledgement entry for a job, and whether it exists.
func (s *AcknowledgementStore) Get(job string) (AcknowledgementEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return e, ok
}

// Lift removes an acknowledgement for the given job.
func (s *AcknowledgementStore) Lift(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all acknowledgement entries.
func (s *AcknowledgementStore) All() []AcknowledgementEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]AcknowledgementEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
