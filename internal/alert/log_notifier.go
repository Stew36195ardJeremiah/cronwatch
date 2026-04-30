package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// LogNotifier sends alerts to a writer (defaults to stdout).
type LogNotifier struct {
	Writer io.Writer
}

// NewLogNotifier creates a LogNotifier writing to the given writer.
// If w is nil, os.Stdout is used.
func NewLogNotifier(w io.Writer) *LogNotifier {
	if w == nil {
		w = os.Stdout
	}
	return &LogNotifier{Writer: w}
}

// Send writes a formatted alert message to the configured writer.
func (l *LogNotifier) Send(level, job, message string) error {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	_, err := fmt.Fprintf(l.Writer, "[%s] [%s] job=%q msg=%q\n", timestamp, level, job, message)
	return err
}
