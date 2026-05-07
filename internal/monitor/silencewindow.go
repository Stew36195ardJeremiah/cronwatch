package monitor

import (
	"sync"
	"time"
)

// SilenceWindow represents a named time window during which alerts are silenced.
type SilenceWindow struct {
	JobName   string
	Start     time.Time
	End       time.Time
	Reason    string
	CreatedAt time.Time
}

// IsActive returns true if the window covers the given time.
func (w SilenceWindow) IsActive(t time.Time) bool {
	return !t.Before(w.Start) && t.Before(w.End)
}

// SilenceWindowStore manages per-job silence windows.
type SilenceWindowStore struct {
	mu      sync.RWMutex
	windows map[string][]SilenceWindow
}

// NewSilenceWindowStore returns an initialised SilenceWindowStore.
func NewSilenceWindowStore() *SilenceWindowStore {
	return &SilenceWindowStore{
		windows: make(map[string][]SilenceWindow),
	}
}

// Add registers a silence window for a job.
func (s *SilenceWindowStore) Add(w SilenceWindow) {
	w.CreatedAt = time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows[w.JobName] = append(s.windows[w.JobName], w)
}

// IsSilenced returns true if any active window covers now for the given job.
func (s *SilenceWindowStore) IsSilenced(job string) bool {
	now := time.Now()
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, w := range s.windows[job] {
		if w.IsActive(now) {
			return true
		}
	}
	return false
}

// Prune removes expired windows for all jobs.
func (s *SilenceWindowStore) Prune() {
	now := time.Now()
	s.mu.Lock()
	defer s.mu.Unlock()
	for job, ws := range s.windows {
		var active []SilenceWindow
		for _, w := range ws {
			if now.Before(w.End) {
				active = append(active, w)
			}
		}
		if len(active) == 0 {
			delete(s.windows, job)
		} else {
			s.windows[job] = active
		}
	}
}

// All returns a snapshot of all windows keyed by job name.
func (s *SilenceWindowStore) All() map[string][]SilenceWindow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]SilenceWindow, len(s.windows))
	for k, v := range s.windows {
		cp := make([]SilenceWindow, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}
