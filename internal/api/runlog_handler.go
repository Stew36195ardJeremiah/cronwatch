package api

import (
	"net/http"

	"github.com/cronwatch/cronwatch/internal/monitor"
)

// RunLogProvider is satisfied by *monitor.RunLog.
type RunLogProvider interface {
	All() []monitor.RunLogEntry
	ForJob(name string) []monitor.RunLogEntry
}

type runLogHandler struct {
	log RunLogProvider
}

func newRunLogHandler(log RunLogProvider) *runLogHandler {
	return &runLogHandler{log: log}
}

// handleRunLog serves GET /runlog[?job=<name>]
// Returns all entries or entries filtered by job name.
func (h *runLogHandler) handleRunLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobName := r.URL.Query().Get("job")

	var entries []monitor.RunLogEntry
	if jobName != "" {
		entries = h.log.ForJob(jobName)
	} else {
		entries = h.log.All()
	}

	type responseEntry struct {
		JobName   string  `json:"job"`
		StartedAt string  `json:"started_at"`
		DurationMs float64 `json:"duration_ms"`
		Success   bool    `json:"success"`
		Message   string  `json:"message,omitempty"`
	}

	result := make([]responseEntry, len(entries))
	for i, e := range entries {
		result[i] = responseEntry{
			JobName:    e.JobName,
			StartedAt:  e.StartedAt.UTC().Format("2006-01-02T15:04:05Z"),
			DurationMs: float64(e.Duration.Milliseconds()),
			Success:    e.Success,
			Message:    e.Message,
		}
	}

	writeJSON(w, http.StatusOK, result)
}
