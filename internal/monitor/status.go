package monitor

import (
	"sync"
	"time"
)

// JobStatus represents the current state of a monitored cron job.
type JobStatus struct {
	Name      string
	LastRun   time.Time
	NextRun   time.Time
	Drift     time.Duration
	Overdue   bool
	FailCount int
	Healthy   bool
}

// StatusStore holds the latest status for all tracked jobs.
type StatusStore struct {
	mu       sync.RWMutex
	statuses map[string]*JobStatus
}

// NewStatusStore creates an empty StatusStore.
func NewStatusStore() *StatusStore {
	return &StatusStore{
		statuses: make(map[string]*JobStatus),
	}
}

// Set updates the status for the given job.
func (s *StatusStore) Set(name string, status JobStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	status.Name = name
	s.statuses[name] = &status
}

// Get retrieves the status for a job by name. Returns false if not found.
func (s *StatusStore) Get(name string) (JobStatus, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	st, ok := s.statuses[name]
	if !ok {
		return JobStatus{}, false
	}
	return *st, true
}

// All returns a snapshot of all job statuses.
func (s *StatusStore) All() []JobStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]JobStatus, 0, len(s.statuses))
	for _, st := range s.statuses {
		result = append(result, *st)
	}
	return result
}

// MarkFailed increments the failure count and marks the job unhealthy.
func (s *StatusStore) MarkFailed(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.statuses[name]
	if !ok {
		s.statuses[name] = &JobStatus{Name: name, FailCount: 1, Healthy: false}
		return
	}
	st.FailCount++
	st.Healthy = false
}

// Reset clears the failure count and marks the job healthy.
func (s *StatusStore) Reset(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.statuses[name]; ok {
		st.FailCount = 0
		st.Healthy = true
	}
}
