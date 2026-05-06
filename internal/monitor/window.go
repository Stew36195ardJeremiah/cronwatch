package monitor

import (
	"sync"
	"time"
)

// WindowEntry holds a single time-bucketed count for a job.
type WindowEntry struct {
	JobName   string
	BucketStart time.Time
	Count     int
}

// WindowStore tracks how many times each job has run within a rolling time window.
type WindowStore struct {
	mu      sync.RWMutex
	window  time.Duration
	buckets map[string][]WindowEntry
}

// NewWindowStore creates a WindowStore with the given rolling window duration.
func NewWindowStore(window time.Duration) *WindowStore {
	if window <= 0 {
		window = time.Hour
	}
	return &WindowStore{
		window:  window,
		buckets: make(map[string][]WindowEntry),
	}
}

// Record adds a run event for the given job at the given time.
func (w *WindowStore) Record(job string, at time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict(job, at)
	entries := w.buckets[job]
	// Increment the current bucket if it shares the same minute, else append.
	if len(entries) > 0 {
		last := &w.buckets[job][len(entries)-1]
		if at.Sub(last.BucketStart) < time.Minute {
			last.Count++
			return
		}
	}
	w.buckets[job] = append(w.buckets[job], WindowEntry{
		JobName:     job,
		BucketStart: at.Truncate(time.Minute),
		Count:       1,
	})
}

// Count returns the total number of runs for a job within the rolling window.
func (w *WindowStore) Count(job string, now time.Time) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.evict(job, now)
	total := 0
	for _, e := range w.buckets[job] {
		total += e.Count
	}
	return total
}

// All returns a snapshot of all current window entries across all jobs.
func (w *WindowStore) All(now time.Time) []WindowEntry {
	w.mu.Lock()
	defer w.mu.Unlock()
	var out []WindowEntry
	for job := range w.buckets {
		w.evict(job, now)
		out = append(out, w.buckets[job]...)
	}
	return out
}

// evict removes entries older than the rolling window. Must be called with lock held.
func (w *WindowStore) evict(job string, now time.Time) {
	cutoff := now.Add(-w.window)
	entries := w.buckets[job]
	i := 0
	for i < len(entries) && entries[i].BucketStart.Before(cutoff) {
		i++
	}
	w.buckets[job] = entries[i:]
}
