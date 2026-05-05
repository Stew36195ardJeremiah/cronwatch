package monitor

import (
	"sync"
	"time"
)

// SuppressionStore tracks alert suppression windows per job.
// If a job is suppressed, alerts will not be dispatched until the window expires.
type SuppressionStore struct {
	mu      sync.RWMutex
	windows map[string]time.Time
}

// NewSuppressionStore creates an empty SuppressionStore.
func NewSuppressionStore() *SuppressionStore {
	return &SuppressionStore{
		windows: make(map[string]time.Time),
	}
}

// Suppress silences alerts for the given job for the specified duration.
func (s *SuppressionStore) Suppress(jobName string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows[jobName] = time.Now().Add(duration)
}

// IsSuppressed reports whether alerts for the given job are currently suppressed.
func (s *SuppressionStore) IsSuppressed(jobName string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	expiry, ok := s.windows[jobName]
	if !ok {
		return false
	}
	return time.Now().Before(expiry)
}

// Lift removes suppression for the given job immediately.
func (s *SuppressionStore) Lift(jobName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.windows, jobName)
}

// All returns a snapshot of all active suppression windows.
func (s *SuppressionStore) All() map[string]time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]time.Time, len(s.windows))
	for k, v := range s.windows {
		out[k] = v
	}
	return out
}
