package alert

import (
	"fmt"
	"log"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Alert holds information about a triggered alert.
type Alert struct {
	JobName   string
	Level     Level
	Message   string
	Timestamp time.Time
}

// Notifier is the interface for sending alerts.
type Notifier interface {
	Send(a Alert) error
}

// LogNotifier sends alerts to the standard logger.
type LogNotifier struct{}

// Send logs the alert to stdout.
func (l *LogNotifier) Send(a Alert) error {
	log.Printf("[%s] [%s] %s — %s\n",
		a.Timestamp.Format(time.RFC3339),
		a.Level,
		a.JobName,
		a.Message,
	)
	return nil
}

// Manager dispatches alerts through one or more Notifiers.
type Manager struct {
	notifiers []Notifier
}

// NewManager creates a Manager with the provided notifiers.
// If none are supplied, a LogNotifier is used by default.
func NewManager(notifiers ...Notifier) *Manager {
	if len(notifiers) == 0 {
		notifiers = []Notifier{&LogNotifier{}}
	}
	return &Manager{notifiers: notifiers}
}

// Warn dispatches a warning-level alert.
func (m *Manager) Warn(jobName, format string, args ...interface{}) {
	m.dispatch(LevelWarn, jobName, fmt.Sprintf(format, args...))
}

// Error dispatches an error-level alert.
func (m *Manager) Error(jobName, format string, args ...interface{}) {
	m.dispatch(LevelError, jobName, fmt.Sprintf(format, args...))
}

func (m *Manager) dispatch(level Level, jobName, message string) {
	a := Alert{
		JobName:   jobName,
		Level:     level,
		Message:   message,
		Timestamp: time.Now().UTC(),
	}
	for _, n := range m.notifiers {
		if err := n.Send(a); err != nil {
			log.Printf("alert notifier error: %v", err)
		}
	}
}
