package monitor

import (
	"log"
	"sync"
	"time"

	"github.com/cronwatch/internal/config"
	"github.com/cronwatch/internal/schedule"
)

// JobStatus tracks the last known execution state of a cron job.
type JobStatus struct {
	Name      string
	LastRun   time.Time
	NextRun   time.Time
	Overdue   bool
	Drift     time.Duration
}

// Monitor watches configured cron jobs and reports drift or failures.
type Monitor struct {
	cfg      *config.Config
	statuses map[string]*JobStatus
	mu       sync.RWMutex
	stopCh   chan struct{}
}

// New creates a new Monitor from the given config.
func New(cfg *config.Config) *Monitor {
	return &Monitor{
		cfg:      cfg,
		statuses: make(map[string]*JobStatus),
		stopCh:   make(chan struct{}),
	}
}

// Start begins polling all configured jobs at the given interval.
func (m *Monitor) Start(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.checkAll()
			case <-m.stopCh:
				return
			}
		}
	}()
}

// Stop signals the monitor to cease polling.
func (m *Monitor) Stop() {
	close(m.stopCh)
}

// RecordRun updates the last run time for a named job.
func (m *Monitor) RecordRun(name string, at time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.statuses[name]; ok {
		s.LastRun = at
		s.Overdue = false
	}
}

// Statuses returns a snapshot of all job statuses.
func (m *Monitor) Statuses() []JobStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make([]JobStatus, 0, len(m.statuses))
	for _, s := range m.statuses {
		out = append(out, *s)
	}
	return out
}

func (m *Monitor) checkAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	for _, job := range m.cfg.Jobs {
		s, ok := m.statuses[job.Name]
		if !ok {
			next, err := schedule.NextRun(job.Schedule, now)
			if err != nil {
				log.Printf("[cronwatch] invalid schedule for job %q: %v", job.Name, err)
				continue
			}
			s = &JobStatus{Name: job.Name, NextRun: next}
			m.statuses[job.Name] = s
		}
		if schedule.IsOverdue(job.Schedule, s.LastRun, now) {
			s.Overdue = true
			s.Drift = schedule.DriftDuration(job.Schedule, s.LastRun, now)
			log.Printf("[cronwatch] ALERT job %q is overdue by %s", job.Name, s.Drift)
		}
	}
}
