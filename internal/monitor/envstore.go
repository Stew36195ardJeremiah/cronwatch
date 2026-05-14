package monitor

import (
	"sync"
	"time"
)

// EnvEntry holds environment metadata captured for a job run.
type EnvEntry struct {
	Job       string            `json:"job"`
	Vars      map[string]string `json:"vars"`
	CapturedAt time.Time        `json:"captured_at"`
}

// EnvStore records environment variable snapshots associated with job runs.
type EnvStore struct {
	mu      sync.RWMutex
	entries map[string]EnvEntry
}

// NewEnvStore initialises an empty EnvStore.
func NewEnvStore() *EnvStore {
	return &EnvStore{
		entries: make(map[string]EnvEntry),
	}
}

// Set records or replaces the environment snapshot for a job.
func (s *EnvStore) Set(job string, vars map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	copy := make(map[string]string, len(vars))
	for k, v := range vars {
		copy[k] = v
	}

	s.entries[job] = EnvEntry{
		Job:        job,
		Vars:       copy,
		CapturedAt: time.Now(),
	}
}

// Get returns the environment snapshot for a job, or false if not found.
func (s *EnvStore) Get(job string) (EnvEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return e, ok
}

// Delete removes the environment snapshot for a job.
func (s *EnvStore) Delete(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all stored environment entries.
func (s *EnvStore) All() []EnvEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]EnvEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
