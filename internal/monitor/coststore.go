package monitor

import (
	"sync"
	"time"
)

// CostEntry holds the recorded execution cost for a single job run.
type CostEntry struct {
	Job       string
	Duration  time.Duration
	RecordedAt time.Time
}

// CostSummary aggregates cost statistics for a job.
type CostSummary struct {
	Job      string
	RunCount int
	TotalMs  int64
	AvgMs    int64
	MaxMs    int64
}

// CostStore tracks cumulative execution cost (duration) per job.
type CostStore struct {
	mu      sync.RWMutex
	entries map[string][]CostEntry
	maxSize int
}

// NewCostStore creates a CostStore with the given per-job sample cap.
func NewCostStore(maxSize int) *CostStore {
	if maxSize <= 0 {
		maxSize = 200
	}
	return &CostStore{
		entries: make(map[string][]CostEntry),
		maxSize: maxSize,
	}
}

// Record appends a new cost entry for the given job.
func (c *CostStore) Record(job string, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := CostEntry{Job: job, Duration: d, RecordedAt: time.Now()}
	c.entries[job] = append(c.entries[job], e)
	if len(c.entries[job]) > c.maxSize {
		c.entries[job] = c.entries[job][1:]
	}
}

// Summary returns aggregated cost statistics for a job.
// Returns false if no data exists for the job.
func (c *CostStore) Summary(job string) (CostSummary, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	samples, ok := c.entries[job]
	if !ok || len(samples) == 0 {
		return CostSummary{}, false
	}
	var total, max int64
	for _, s := range samples {
		ms := s.Duration.Milliseconds()
		total += ms
		if ms > max {
			max = ms
		}
	}
	return CostSummary{
		Job:      job,
		RunCount: len(samples),
		TotalMs:  total,
		AvgMs:    total / int64(len(samples)),
		MaxMs:    max,
	}, true
}

// All returns a snapshot of summaries for every tracked job.
func (c *CostStore) All() []CostSummary {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]CostSummary, 0, len(c.entries))
	for job, samples := range c.entries {
		if len(samples) == 0 {
			continue
		}
		var total, max int64
		for _, s := range samples {
			ms := s.Duration.Milliseconds()
			total += ms
			if ms > max {
				max = ms
			}
		}
		out = append(out, CostSummary{
			Job:      job,
			RunCount: len(samples),
			TotalMs:  total,
			AvgMs:    total / int64(len(samples)),
			MaxMs:    max,
		})
	}
	return out
}
