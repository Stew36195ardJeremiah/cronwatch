package monitor

import (
	"sync"
	"time"
)

// EscalationPolicy defines thresholds for escalating alerts.
type EscalationPolicy struct {
	JobName       string
	WarnAfter     time.Duration
	CriticalAfter time.Duration
}

// EscalationLevel represents the severity of an escalation.
type EscalationLevel int

const (
	EscalationNone     EscalationLevel = iota
	EscalationWarn
	EscalationCritical
)

func (l EscalationLevel) String() string {
	switch l {
	case EscalationWarn:
		return "warn"
	case EscalationCritical:
		return "critical"
	default:
		return "none"
	}
}

// EscalationStore tracks per-job escalation state.
type EscalationStore struct {
	mu       sync.RWMutex
	policies map[string]EscalationPolicy
	triggered map[string]time.Time // when the issue was first detected
}

// NewEscalationStore creates an empty EscalationStore.
func NewEscalationStore() *EscalationStore {
	return &EscalationStore{
		policies:  make(map[string]EscalationPolicy),
		triggered: make(map[string]time.Time),
	}
}

// SetPolicy registers or replaces an escalation policy for a job.
func (s *EscalationStore) SetPolicy(p EscalationPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[p.JobName] = p
}

// Trigger records the first time an issue was detected for a job.
// Subsequent calls before Reset are no-ops.
func (s *EscalationStore) Trigger(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.triggered[job]; !exists {
		s.triggered[job] = time.Now()
	}
}

// Reset clears the triggered state for a job (e.g. after recovery).
func (s *EscalationStore) Reset(job string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.triggered, job)
}

// Level returns the current escalation level for a job based on elapsed time.
func (s *EscalationStore) Level(job string) EscalationLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, hasPolicy := s.policies[job]
	if !hasPolicy {
		return EscalationNone
	}
	triggeredAt, active := s.triggered[job]
	if !active {
		return EscalationNone
	}
	elapsed := time.Since(triggeredAt)
	if policy.CriticalAfter > 0 && elapsed >= policy.CriticalAfter {
		return EscalationCritical
	}
	if policy.WarnAfter > 0 && elapsed >= policy.WarnAfter {
		return EscalationWarn
	}
	return EscalationNone
}

// All returns a snapshot of all active escalation levels.
func (s *EscalationStore) All() map[string]EscalationLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]EscalationLevel, len(s.triggered))
	for job := range s.triggered {
		out[job] = s.Level(job)
	}
	return out
}
