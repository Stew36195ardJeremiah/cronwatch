package monitor

import (
	"sync"
	"time"
)

// DeadLetterEntry represents a failed alert that could not be delivered.
type DeadLetterEntry struct {
	JobName   string
	Level     string
	Message   string
	Notifier  string
	Err       string
	FailedAt  time.Time
	Attempts  int
}

// DeadLetterStore holds alerts that failed to dispatch to any notifier.
type DeadLetterStore struct {
	mu      sync.RWMutex
	entries []DeadLetterEntry
	maxSize int
}

const defaultDeadLetterMax = 200

// NewDeadLetterStore creates a DeadLetterStore with an optional max size.
func NewDeadLetterStore(maxSize int) *DeadLetterStore {
	if maxSize <= 0 {
		maxSize = defaultDeadLetterMax
	}
	return &DeadLetterStore{maxSize: maxSize}
}

// Record appends a failed alert entry, evicting the oldest if at capacity.
func (d *DeadLetterStore) Record(entry DeadLetterEntry) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.entries) >= d.maxSize {
		d.entries = d.entries[1:]
	}
	d.entries = append(d.entries, entry)
}

// All returns a snapshot of all dead-letter entries.
func (d *DeadLetterStore) All() []DeadLetterEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]DeadLetterEntry, len(d.entries))
	copy(out, d.entries)
	return out
}

// ForJob returns dead-letter entries for a specific job.
func (d *DeadLetterStore) ForJob(jobName string) []DeadLetterEntry {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var out []DeadLetterEntry
	for _, e := range d.entries {
		if e.JobName == jobName {
			out = append(out, e)
		}
	}
	return out
}

// Clear removes all entries from the store.
func (d *DeadLetterStore) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.entries = nil
}

// Len returns the current number of entries.
func (d *DeadLetterStore) Len() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.entries)
}
