package monitor

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker for a job.
type CircuitState int

const (
	CircuitClosed CircuitState = iota // normal operation
	CircuitOpen                        // alerting suppressed, job considered failing
	CircuitHalfOpen                    // trial period, next success closes circuit
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type circuitEntry struct {
	state      CircuitState
	failures   int
	openedAt   time.Time
	recoverIn  time.Duration
}

// CircuitBreakerStore tracks per-job circuit breaker state.
type CircuitBreakerStore struct {
	mu          sync.RWMutex
	entries     map[string]*circuitEntry
	maxFailures int
	recoverIn   time.Duration
}

// NewCircuitBreakerStore creates a store with the given failure threshold and
// recovery window after which an open circuit transitions to half-open.
func NewCircuitBreakerStore(maxFailures int, recoverIn time.Duration) *CircuitBreakerStore {
	if maxFailures <= 0 {
		maxFailures = 3
	}
	if recoverIn <= 0 {
		recoverIn = 5 * time.Minute
	}
	return &CircuitBreakerStore{
		entries:     make(map[string]*circuitEntry),
		maxFailures: maxFailures,
		recoverIn:   recoverIn,
	}
}

// RecordFailure increments the failure count for a job and may open the circuit.
func (c *CircuitBreakerStore) RecordFailure(job string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.getOrCreate(job)
	e.failures++
	if e.state == CircuitClosed && e.failures >= c.maxFailures {
		e.state = CircuitOpen
		e.openedAt = time.Now()
	}
}

// RecordSuccess resets failure count; closes a half-open circuit.
func (c *CircuitBreakerStore) RecordSuccess(job string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.getOrCreate(job)
	e.failures = 0
	if e.state == CircuitHalfOpen || e.state == CircuitOpen {
		e.state = CircuitClosed
	}
}

// State returns the current (possibly transitioned) circuit state for a job.
func (c *CircuitBreakerStore) State(job string) CircuitState {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := c.getOrCreate(job)
	if e.state == CircuitOpen && time.Since(e.openedAt) >= c.recoverIn {
		e.state = CircuitHalfOpen
	}
	return e.state
}

// Reset clears all state for a job.
func (c *CircuitBreakerStore) Reset(job string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, job)
}

// All returns a snapshot of all job circuit states.
func (c *CircuitBreakerStore) All() map[string]CircuitState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]CircuitState, len(c.entries))
	for k, v := range c.entries {
		out[k] = v.state
	}
	return out
}

func (c *CircuitBreakerStore) getOrCreate(job string) *circuitEntry {
	e, ok := c.entries[job]
	if !ok {
		e = &circuitEntry{state: CircuitClosed, recoverIn: c.recoverIn}
		c.entries[job] = e
	}
	return e
}
