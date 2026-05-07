package alert

import (
	"fmt"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Alert holds the details of a single alert event.
type Alert struct {
	Job     string
	Message string
	Level   Level
	Time    time.Time
}

// Notifier is the interface implemented by all alert backends.
type Notifier interface {
	Send(a Alert) error
}

// Manager dispatches alerts to one or more notifiers.
type Manager struct {
	notifiers []Notifier
}

// NewManager creates a Manager. If no notifiers are provided, a
// default LogNotifier writing to stdout is added.
func NewManager(notifiers ...Notifier) *Manager {
	if len(notifiers) == 0 {
		notifiers = []Notifier{NewLogNotifier(nil)}
	}
	return &Manager{notifiers: notifiers}
}

// AddNotifier appends a notifier to the manager.
func (m *Manager) AddNotifier(n Notifier) {
	m.notifiers = append(m.notifiers, n)
}

// Warn dispatches a warning-level alert for the named job.
func (m *Manager) Warn(job, message string) {
	m.dispatch(Alert{Job: job, Message: message, Level: LevelWarn, Time: time.Now()})
}

// Error dispatches an error-level alert for the named job.
func (m *Manager) Error(job, message string) {
	m.dispatch(Alert{Job: job, Message: message, Level: LevelError, Time: time.Now()})
}

// Warnf dispatches a warning-level alert for the named job using a
// printf-style format string.
func (m *Manager) Warnf(job, format string, args ...any) {
	m.Warn(job, fmt.Sprintf(format, args...))
}

// Errorf dispatches an error-level alert for the named job using a
// printf-style format string.
func (m *Manager) Errorf(job, format string, args ...any) {
	m.Error(job, fmt.Sprintf(format, args...))
}

// dispatch sends the alert to every registered notifier, logging failures.
func (m *Manager) dispatch(a Alert) {
	for _, n := range m.notifiers {
		if err := n.Send(a); err != nil {
			fmt.Printf("cronwatch: alert dispatch error: %v\n", err)
		}
	}
}
