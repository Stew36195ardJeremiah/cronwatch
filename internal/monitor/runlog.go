package monitor

import (
	"sync"
	"time"
)

// RunLogEntry represents a single recorded execution of a cron job.
type RunLogEntry struct {
	JobName   string
	StartedAt time.Time
	Duration  time.Duration
	Success   bool
	Message   string
}

// RunLog stores a bounded, ordered log of recent job executions across all jobs.
type RunLog struct {
	mu      sync.RWMutex
	entries []RunLogEntry
	maxSize int
}

// NewRunLog creates a RunLog with the given maximum capacity.
func NewRunLog(maxSize int) *RunLog {
	if maxSize <= 0 {
		maxSize = 200
	}
	return &RunLog{
		entries: make([]RunLogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Append adds a new entry to the run log, evicting the oldest if at capacity.
func (r *RunLog) Append(entry RunLogEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.entries) >= r.maxSize {
		r.entries = r.entries[1:]
	}
	r.entries = append(r.entries, entry)
}

// All returns a snapshot of all entries, newest last.
func (r *RunLog) All() []RunLogEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	snap := make([]RunLogEntry, len(r.entries))
	copy(snap, r.entries)
	return snap
}

// ForJob returns all entries for a specific job name.
func (r *RunLog) ForJob(name string) []RunLogEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []RunLogEntry
	for _, e := range r.entries {
		if e.JobName == name {
			result = append(result, e)
		}
	}
	return result
}

// Len returns the current number of entries.
func (r *RunLog) Len() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.entries)
}
