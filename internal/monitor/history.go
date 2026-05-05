package monitor

import (
	"sync"
	"time"
)

// RunRecord captures a single execution record for a cron job.
type RunRecord struct {
	JobName   string
	StartedAt time.Time
	Drift     time.Duration
	Overdue   bool
}

// History maintains a bounded ring-buffer of run records per job.
type History struct {
	mu      sync.RWMutex
	records map[string][]RunRecord
	maxSize int
}

// NewHistory creates a History that keeps at most maxSize records per job.
func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 50
	}
	return &History{
		records: make(map[string][]RunRecord),
		maxSize: maxSize,
	}
}

// Record appends a RunRecord for the given job, evicting the oldest if needed.
func (h *History) Record(r RunRecord) {
	h.mu.Lock()
	defer h.mu.Unlock()

	buf := h.records[r.JobName]
	if len(buf) >= h.maxSize {
		buf = buf[1:]
	}
	h.records[r.JobName] = append(buf, r)
}

// Get returns a copy of all records for the given job.
func (h *History) Get(jobName string) []RunRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	src := h.records[jobName]
	out := make([]RunRecord, len(src))
	copy(out, src)
	return out
}

// All returns a snapshot of every record across all jobs.
func (h *History) All() map[string][]RunRecord {
	h.mu.RLock()
	defer h.mu.RUnlock()

	out := make(map[string][]RunRecord, len(h.records))
	for k, v := range h.records {
		cp := make([]RunRecord, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

// Clear removes all records for a specific job.
func (h *History) Clear(jobName string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.records, jobName)
}
