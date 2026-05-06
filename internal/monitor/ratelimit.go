package monitor

import (
	"sync"
	"time"
)

// RateLimitEntry tracks alert rate limiting for a single job.
type RateLimitEntry struct {
	LastAlertAt time.Time
	Count       int
}

// RateLimitStore prevents alert storms by throttling repeated alerts per job.
type RateLimitStore struct {
	mu       sync.RWMutex
	entries  map[string]*RateLimitEntry
	window   time.Duration
	maxCount int
}

// NewRateLimitStore creates a store that allows at most maxCount alerts per job
// within the given window duration.
func NewRateLimitStore(window time.Duration, maxCount int) *RateLimitStore {
	if maxCount <= 0 {
		maxCount = 3
	}
	if window <= 0 {
		window = 10 * time.Minute
	}
	return &RateLimitStore{
		entries:  make(map[string]*RateLimitEntry),
		window:   window,
		maxCount: maxCount,
	}
}

// Allow returns true if an alert for the given job should be dispatched.
// It increments the counter and resets it when the window expires.
func (r *RateLimitStore) Allow(job string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	e, ok := r.entries[job]
	if !ok || now.Sub(e.LastAlertAt) > r.window {
		r.entries[job] = &RateLimitEntry{LastAlertAt: now, Count: 1}
		return true
	}
	if e.Count >= r.maxCount {
		return false
	}
	e.Count++
	e.LastAlertAt = now
	return true
}

// Reset clears the rate limit state for a specific job.
func (r *RateLimitStore) Reset(job string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.entries, job)
}

// All returns a snapshot of current rate limit entries.
func (r *RateLimitStore) All() map[string]RateLimitEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make(map[string]RateLimitEntry, len(r.entries))
	for k, v := range r.entries {
		out[k] = *v
	}
	return out
}
