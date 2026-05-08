package monitor

import (
	"sync"
	"time"
)

// QuotaEntry tracks alert emission counts within a rolling window.
type QuotaEntry struct {
	Job       string
	Count     int
	Limit     int
	WindowDur time.Duration
	WindowEnd time.Time
}

// QuotaStore enforces per-job alert quotas within time windows.
type QuotaStore struct {
	mu      sync.RWMutex
	entries map[string]*QuotaEntry
	defaultLimit  int
	defaultWindow time.Duration
}

// NewQuotaStore creates a QuotaStore with the given default limit and window.
func NewQuotaStore(defaultLimit int, defaultWindow time.Duration) *QuotaStore {
	return &QuotaStore{
		entries:       make(map[string]*QuotaEntry),
		defaultLimit:  defaultLimit,
		defaultWindow: defaultWindow,
	}
}

// SetLimit overrides the quota limit and window for a specific job.
func (q *QuotaStore) SetLimit(job string, limit int, window time.Duration) {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now()
	q.entries[job] = &QuotaEntry{
		Job:       job,
		Limit:     limit,
		WindowDur: window,
		WindowEnd: now.Add(window),
	}
}

// Allow returns true if the job is within its quota, incrementing the counter.
func (q *QuotaStore) Allow(job string) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	now := time.Now()
	e, ok := q.entries[job]
	if !ok {
		e = &QuotaEntry{
			Job:       job,
			Limit:     q.defaultLimit,
			WindowDur: q.defaultWindow,
			WindowEnd: now.Add(q.defaultWindow),
		}
		q.entries[job] = e
	}
	if now.After(e.WindowEnd) {
		e.Count = 0
		e.WindowEnd = now.Add(e.WindowDur)
	}
	if e.Count >= e.Limit {
		return false
	}
	e.Count++
	return true
}

// Get returns the current QuotaEntry for a job, or nil if unknown.
func (q *QuotaStore) Get(job string) *QuotaEntry {
	q.mu.RLock()
	defer q.mu.RUnlock()
	e, ok := q.entries[job]
	if !ok {
		return nil
	}
	copy := *e
	return &copy
}

// Reset clears the counter for a job without removing its limit config.
func (q *QuotaStore) Reset(job string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if e, ok := q.entries[job]; ok {
		e.Count = 0
		e.WindowEnd = time.Now().Add(e.WindowDur)
	}
}

// All returns a snapshot of all quota entries.
func (q *QuotaStore) All() []QuotaEntry {
	q.mu.RLock()
	defer q.mu.RUnlock()
	out := make([]QuotaEntry, 0, len(q.entries))
	for _, e := range q.entries {
		out = append(out, *e)
	}
	return out
}
