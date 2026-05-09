package monitor

import (
	"sync"
	"time"
)

// OwnerEntry holds ownership metadata for a cron job.
type OwnerEntry struct {
	Job       string    `json:"job"`
	Owner     string    `json:"owner"`
	Team      string    `json:"team"`
	Contact   string    `json:"contact"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OwnershipStore tracks ownership assignments for cron jobs.
type OwnershipStore struct {
	mu      sync.RWMutex
	entries map[string]OwnerEntry
}

// NewOwnershipStore creates an empty OwnershipStore.
func NewOwnershipStore() *OwnershipStore {
	return &OwnershipStore{
		entries: make(map[string]OwnerEntry),
	}
}

// Set assigns ownership for the given job.
func (s *OwnershipStore) Set(job, owner, team, contact string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[job] = OwnerEntry{
		Job:       job,
		Owner:     owner,
		Team:      team,
		Contact:   contact,
		UpdatedAt: time.Now(),
	}
}

// Get returns the ownership entry for the given job, if it exists.
func (s *OwnershipStore) Get(job string) (OwnerEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.entries[job]
	return e, ok
}

// Remove deletes the ownership record for a job.
func (s *OwnershipStore) Remove(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, job)
}

// All returns a snapshot of all ownership entries.
func (s *OwnershipStore) All() []OwnerEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]OwnerEntry, 0, len(s.entries))
	for _, e := range s.entries {
		out = append(out, e)
	}
	return out
}
