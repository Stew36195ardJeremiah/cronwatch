package monitor

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"sync"
	"time"
)

// FingerprintEntry records a deduplicated alert event.
type FingerprintEntry struct {
	JobName   string
	Level     string
	Message   string
	Hash      string
	FirstSeen time.Time
	LastSeen  time.Time
	Count     int
}

// FingerprintStore deduplicates alerts by hashing job+level+message.
type FingerprintStore struct {
	mu      sync.RWMutex
	entries map[string]*FingerprintEntry
	ttl     time.Duration
}

// NewFingerprintStore creates a store that expires entries after ttl.
func NewFingerprintStore(ttl time.Duration) *FingerprintStore {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &FingerprintStore{
		entries: make(map[string]*FingerprintEntry),
		ttl:     ttl,
	}
}

func fingerprintHash(job, level, message string) string {
	h := sha256.Sum256([]byte(job + "|" + level + "|" + message))
	return fmt.Sprintf("%x", h[:8])
}

// Record records an alert event; returns true if this is a new (non-duplicate) fingerprint.
func (f *FingerprintStore) Record(job, level, message string) (isNew bool) {
	hash := fingerprintHash(job, level, message)
	now := time.Now()

	f.mu.Lock()
	defer f.mu.Unlock()

	f.evictExpired(now)

	if e, ok := f.entries[hash]; ok {
		e.LastSeen = now
		e.Count++
		return false
	}

	f.entries[hash] = &FingerprintEntry{
		JobName:   job,
		Level:     level,
		Message:   message,
		Hash:      hash,
		FirstSeen: now,
		LastSeen:  now,
		Count:     1,
	}
	return true
}

// All returns a snapshot of all active fingerprint entries sorted by first seen.
func (f *FingerprintStore) All() []FingerprintEntry {
	f.mu.RLock()
	defer f.mu.RUnlock()

	out := make([]FingerprintEntry, 0, len(f.entries))
	for _, e := range f.entries {
		out = append(out, *e)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].FirstSeen.Before(out[j].FirstSeen)
	})
	return out
}

// evictExpired removes entries older than ttl. Must be called with lock held.
func (f *FingerprintStore) evictExpired(now time.Time) {
	for hash, e := range f.entries {
		if now.Sub(e.LastSeen) > f.ttl {
			delete(f.entries, hash)
		}
	}
}
