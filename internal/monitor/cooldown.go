package monitor

import (
	"sync"
	"time"
)

// CooldownStore tracks per-job cooldown periods to prevent alert storms
// after a job recovers. During cooldown, alerts are suppressed even if
// the job briefly appears overdue again.
type CooldownStore struct {
	mu      sync.Mutex
	entries map[string]cooldownEntry
	defaultDuration time.Duration
}

type cooldownEntry struct {
	activatedAt time.Time
	duration    time.Duration
}

// NewCooldownStore creates a CooldownStore with the given default cooldown duration.
func NewCooldownStore(defaultDuration time.Duration) *CooldownStore {
	if defaultDuration <= 0 {
		defaultDuration = 5 * time.Minute
	}
	return &CooldownStore{
		entries:         make(map[string]cooldownEntry),
		defaultDuration: defaultDuration,
	}
}

// Activate starts a cooldown period for the given job using the default duration.
func (c *CooldownStore) Activate(job string) {
	c.ActivateFor(job, c.defaultDuration)
}

// ActivateFor starts a cooldown period for the given job with a specific duration.
func (c *CooldownStore) ActivateFor(job string, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[job] = cooldownEntry{
		activatedAt: time.Now(),
		duration:    d,
	}
}

// InCooldown reports whether the given job is currently in a cooldown period.
func (c *CooldownStore) InCooldown(job string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[job]
	if !ok {
		return false
	}
	if time.Since(e.activatedAt) >= e.duration {
		delete(c.entries, job)
		return false
	}
	return true
}

// Lift removes any active cooldown for the given job immediately.
func (c *CooldownStore) Lift(job string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, job)
}

// All returns a snapshot of all active cooldown entries as a map of
// job name to remaining duration.
func (c *CooldownStore) All() map[string]time.Duration {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make(map[string]time.Duration, len(c.entries))
	for job, e := range c.entries {
		remaining := e.duration - time.Since(e.activatedAt)
		if remaining > 0 {
			out[job] = remaining
		}
	}
	return out
}
