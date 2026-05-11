package monitor

import (
	"sync"
	"time"
)

// AuditAction represents the type of action recorded in the audit log.
type AuditAction string

const (
	AuditActionSuppress   AuditAction = "suppress"
	AuditActionLift       AuditAction = "lift"
	AuditActionPause      AuditAction = "pause"
	AuditActionResume     AuditAction = "resume"
	AuditActionAcknowledge AuditAction = "acknowledge"
	AuditActionSetPolicy  AuditAction = "set_policy"
	AuditActionDeleteRoute AuditAction = "delete_route"
)

// AuditEntry records a single administrative action taken against a job.
type AuditEntry struct {
	Timestamp time.Time   `json:"timestamp"`
	Job       string      `json:"job"`
	Action    AuditAction `json:"action"`
	Actor     string      `json:"actor"`
	Detail    string      `json:"detail,omitempty"`
}

// AuditLogStore retains a bounded history of audit entries.
type AuditLogStore struct {
	mu      sync.Mutex
	entries []AuditEntry
	maxSize int
}

const defaultAuditMaxSize = 500

// NewAuditLogStore creates an AuditLogStore with the default capacity.
func NewAuditLogStore() *AuditLogStore {
	return &AuditLogStore{maxSize: defaultAuditMaxSize}
}

// Record appends an audit entry, evicting the oldest if capacity is exceeded.
func (s *AuditLogStore) Record(job string, action AuditAction, actor, detail string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := AuditEntry{
		Timestamp: time.Now().UTC(),
		Job:       job,
		Action:    action,
		Actor:     actor,
		Detail:    detail,
	}
	if len(s.entries) >= s.maxSize {
		s.entries = s.entries[1:]
	}
	s.entries = append(s.entries, entry)
}

// All returns a snapshot of all audit entries.
func (s *AuditLogStore) All() []AuditEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]AuditEntry, len(s.entries))
	copy(out, s.entries)
	return out
}

// ForJob returns audit entries filtered to a specific job.
func (s *AuditLogStore) ForJob(job string) []AuditEntry {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []AuditEntry
	for _, e := range s.entries {
		if e.Job == job {
			out = append(out, e)
		}
	}
	return out
}
