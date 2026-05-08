package monitor

import (
	"math"
	"sync"
	"time"
)

// TrendEntry holds computed trend data for a single job.
type TrendEntry struct {
	Job        string        `json:"job"`
	SampleCount int          `json:"sample_count"`
	AvgDrift   time.Duration `json:"avg_drift_ns"`
	StdDev     time.Duration `json:"std_dev_ns"`
	Trending   string        `json:"trending"` // "stable", "increasing", "decreasing"
	UpdatedAt  time.Time     `json:"updated_at"`
}

// TrendStore computes drift trend statistics per job.
type TrendStore struct {
	mu      sync.RWMutex
	samples map[string][]time.Duration
	max     int
}

// NewTrendStore creates a TrendStore with a maximum sample window.
func NewTrendStore(maxSamples int) *TrendStore {
	if maxSamples <= 0 {
		maxSamples = 50
	}
	return &TrendStore{
		samples: make(map[string][]time.Duration),
		max:     maxSamples,
	}
}

// Record adds a drift sample for the given job.
func (t *TrendStore) Record(job string, drift time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := t.samples[job]
	s = append(s, drift)
	if len(s) > t.max {
		s = s[len(s)-t.max:]
	}
	t.samples[job] = s
}

// Get returns computed trend statistics for a job.
func (t *TrendStore) Get(job string) (TrendEntry, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s, ok := t.samples[job]
	if !ok || len(s) == 0 {
		return TrendEntry{}, false
	}
	return computeTrend(job, s), true
}

// All returns trend entries for all tracked jobs.
func (t *TrendStore) All() []TrendEntry {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]TrendEntry, 0, len(t.samples))
	for job, s := range t.samples {
		if len(s) > 0 {
			out = append(out, computeTrend(job, s))
		}
	}
	return out
}

func computeTrend(job string, samples []time.Duration) TrendEntry {
	n := float64(len(samples))
	var sum float64
	for _, d := range samples {
		sum += float64(d)
	}
	avg := sum / n

	var variance float64
	for _, d := range samples {
		diff := float64(d) - avg
		variance += diff * diff
	}
	if n > 1 {
		variance /= n - 1
	}
	stddev := math.Sqrt(variance)

	trending := "stable"
	if len(samples) >= 4 {
		mid := len(samples) / 2
		var early, late float64
		for _, d := range samples[:mid] {
			early += float64(d)
		}
		for _, d := range samples[mid:] {
			late += float64(d)
		}
		early /= float64(mid)
		late /= float64(len(samples) - mid)
		threshold := avg * 0.1
		if late-early > threshold {
			trending = "increasing"
		} else if early-late > threshold {
			trending = "decreasing"
		}
	}

	return TrendEntry{
		Job:        job,
		SampleCount: len(samples),
		AvgDrift:   time.Duration(avg),
		StdDev:     time.Duration(stddev),
		Trending:   trending,
		UpdatedAt:  time.Now(),
	}
}
