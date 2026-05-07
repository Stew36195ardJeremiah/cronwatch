package monitor

import (
	"fmt"
	"sync"
)

// RouteRule defines a mapping from a job name pattern to a named alert channel.
type RouteRule struct {
	Job     string // exact job name or "*" wildcard
	Channel string // named alert channel (e.g. "slack", "pagerduty")
}

// RoutingStore maps jobs to alert channel names.
type RoutingStore struct {
	mu    sync.RWMutex
	routes map[string]string // job -> channel
	defaultChannel string
}

// NewRoutingStore creates a new RoutingStore with an optional default channel.
func NewRoutingStore(defaultChannel string) *RoutingStore {
	return &RoutingStore{
		routes:         make(map[string]string),
		defaultChannel: defaultChannel,
	}
}

// Set assigns a channel to a specific job.
func (r *RoutingStore) Set(job, channel string) error {
	if job == "" {
		return fmt.Errorf("job name must not be empty")
	}
	if channel == "" {
		return fmt.Errorf("channel must not be empty")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[job] = channel
	return nil
}

// Resolve returns the channel for a job, falling back to the default.
func (r *RoutingStore) Resolve(job string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if ch, ok := r.routes[job]; ok {
		return ch
	}
	return r.defaultChannel
}

// Remove deletes a routing rule for a job.
func (r *RoutingStore) Remove(job string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.routes, job)
}

// All returns a snapshot of all routing rules.
func (r *RoutingStore) All() []RouteRule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]RouteRule, 0, len(r.routes))
	for job, ch := range r.routes {
		out = append(out, RouteRule{Job: job, Channel: ch})
	}
	return out
}

// SetDefault updates the fallback channel.
func (r *RoutingStore) SetDefault(channel string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defaultChannel = channel
}

// Default returns the current fallback channel.
func (r *RoutingStore) Default() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.defaultChannel
}
