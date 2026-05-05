package monitor

import (
	"sync"
	"time"
)

// RetryPolicy defines how many times an alert should be retried before giving up.
type RetryPolicy struct {
	MaxAttempts int
	Interval    time.Duration
}

// RetryState tracks the current retry state for a single job.
type RetryState struct {
	Attempts  int
	LastRetry time.Time
	Exhausted bool
}

// RetryStore manages per-job retry state.
type RetryStore struct {
	mu     sync.RWMutex
	states map[string]*RetryState
	policy RetryPolicy
}

// NewRetryStore creates a RetryStore with the given policy.
func NewRetryStore(policy RetryPolicy) *RetryStore {
	if policy.MaxAttempts <= 0 {
		policy.MaxAttempts = 3
	}
	if policy.Interval <= 0 {
		policy.Interval = 5 * time.Minute
	}
	return &RetryStore{
		states: make(map[string]*RetryState),
		policy: policy,
	}
}

// ShouldRetry returns true if the job should fire another alert attempt.
func (r *RetryStore) ShouldRetry(job string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, ok := r.states[job]
	if !ok {
		r.states[job] = &RetryState{Attempts: 1, LastRetry: time.Now()}
		return true
	}
	if state.Exhausted {
		return false
	}
	if time.Since(state.LastRetry) < r.policy.Interval {
		return false
	}
	state.Attempts++
	state.LastRetry = time.Now()
	if state.Attempts >= r.policy.MaxAttempts {
		state.Exhausted = true
	}
	return true
}

// Reset clears the retry state for a job (e.g. after a successful run).
func (r *RetryStore) Reset(job string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.states, job)
}

// State returns a copy of the retry state for a job.
func (r *RetryStore) State(job string) (RetryState, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.states[job]
	if !ok {
		return RetryState{}, false
	}
	return *s, true
}
