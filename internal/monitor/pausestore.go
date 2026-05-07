package monitor

import (
	"sync"
	"time"
)

// PauseEntry holds the pause state for a single job.
type PauseEntry struct {
	JobName   string    `json:"job_name"`
	PausedAt  time.Time `json:"paused_at"`
	Reason    string    `json:"reason,omitempty"`
	PausedBy  string    `json:"paused_by,omitempty"`
}

// PauseStore tracks which jobs are manually paused (checks skipped).
type PauseStore struct {
	mu      sync.RWMutex
	entries map[string]PauseEntry
}

// NewPauseStore returns an empty PauseStore.
func NewPauseStore() *PauseStore {
	return &PauseStore{
		entries: make(map[string]PauseEntry),
	}
}

// Pause marks a job as paused.
func (s *PauseStore) Pause(jobName, reason, pausedBy string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[jobName] = PauseEntry{
		JobName:  jobName,
		PausedAt: time.Now(),
		Reason:   reason,
		PausedBy: pausedBy,
	}
}

// Resume removes a job's paused state.
func (s *PauseStore) Resume(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, jobName)
}

// IsPaused returns true if the job is currently paused.
func (s *PauseStore) IsPaused(jobName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.entries[jobName]
	return ok
}

// Get returns the PauseEntry for a job, or false if not paused.
func (s *PauseStore) Get(jobName string) (PauseEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[jobName]
	return e, ok
}

// All returns a snapshot of all paused jobs.
func (s *PauseStore) All() []PauseEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]PauseEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
