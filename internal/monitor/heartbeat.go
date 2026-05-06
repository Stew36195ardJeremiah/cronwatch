package monitor

import (
	"sync"
	"time"
)

// HeartbeatRecord holds the last known heartbeat time for a job.
type HeartbeatRecord struct {
	JobName   string
	LastSeen  time.Time
	TTL       time.Duration
	Expired   bool
}

// HeartbeatStore tracks periodic heartbeat signals from jobs and
// exposes whether a job's heartbeat has gone silent.
type HeartbeatStore struct {
	mu      sync.RWMutex
	records map[string]HeartbeatRecord
}

// NewHeartbeatStore returns an initialised HeartbeatStore.
func NewHeartbeatStore() *HeartbeatStore {
	return &HeartbeatStore{
		records: make(map[string]HeartbeatRecord),
	}
}

// Beat registers a heartbeat for the given job with the supplied TTL.
// Calling Beat resets the expiry window.
func (h *HeartbeatStore) Beat(job string, ttl time.Duration) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.records[job] = HeartbeatRecord{
		JobName:  job,
		LastSeen: time.Now(),
		TTL:      ttl,
		Expired:  false,
	}
}

// IsExpired returns true when the job's last heartbeat is older than its TTL,
// or when no heartbeat has ever been recorded for the job.
func (h *HeartbeatStore) IsExpired(job string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	rec, ok := h.records[job]
	if !ok {
		return true
	}
	return time.Since(rec.LastSeen) > rec.TTL
}

// Get returns the HeartbeatRecord for a job and whether it exists.
func (h *HeartbeatStore) Get(job string) (HeartbeatRecord, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	rec, ok := h.records[job]
	if ok {
		rec.Expired = time.Since(rec.LastSeen) > rec.TTL
	}
	return rec, ok
}

// All returns a snapshot of all heartbeat records with up-to-date Expired flags.
func (h *HeartbeatStore) All() []HeartbeatRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]HeartbeatRecord, 0, len(h.records))
	for _, rec := range h.records {
		rec.Expired = time.Since(rec.LastSeen) > rec.TTL
		out = append(out, rec)
	}
	return out
}
