package monitor

import (
	"sync"
	"time"
)

// JobMetrics holds aggregated runtime statistics for a single job.
type JobMetrics struct {
	JobName     string
	RunCount    int
	FailCount   int
	AvgDrift    time.Duration
	MaxDrift    time.Duration
	LastDrift   time.Duration
	LastUpdated time.Time
}

// MetricsStore accumulates per-job metrics over time.
type MetricsStore struct {
	mu   sync.RWMutex
	data map[string]*JobMetrics
}

// NewMetricsStore returns an initialised MetricsStore.
func NewMetricsStore() *MetricsStore {
	return &MetricsStore{
		data: make(map[string]*JobMetrics),
	}
}

// Record updates metrics for jobName using the supplied drift value.
// A negative drift means the job ran early; positive means late.
func (m *MetricsStore) Record(jobName string, drift time.Duration, failed bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	met, ok := m.data[jobName]
	if !ok {
		met = &JobMetrics{JobName: jobName}
		m.data[jobName] = met
	}

	met.RunCount++
	if failed {
		met.FailCount++
	}

	abs := drift
	if abs < 0 {
		abs = -abs
	}

	met.LastDrift = drift
	if abs > met.MaxDrift {
		met.MaxDrift = abs
	}

	// rolling average of absolute drift
	met.AvgDrift = time.Duration((int64(met.AvgDrift)*int64(met.RunCount-1) + int64(abs)) / int64(met.RunCount))
	met.LastUpdated = time.Now()
}

// Get returns a copy of the metrics for jobName and whether it was found.
func (m *MetricsStore) Get(jobName string) (JobMetrics, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	met, ok := m.data[jobName]
	if !ok {
		return JobMetrics{}, false
	}
	return *met, true
}

// All returns a snapshot of metrics for every tracked job.
func (m *MetricsStore) All() []JobMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]JobMetrics, 0, len(m.data))
	for _, v := range m.data {
		out = append(out, *v)
	}
	return out
}
