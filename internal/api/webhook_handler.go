package api

import (
	"encoding/json"
	"net/http"
	"time"
)

// RunEvent represents an inbound webhook payload from a cron job.
type RunEvent struct {
	Job       string    `json:"job"`
	Status    string    `json:"status"` // "success" | "failure"
	Timestamp time.Time `json:"timestamp"`
}

// handleIngestRun accepts POST /ingest and records a run event.
func (s *Server) handleIngestRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var event RunEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if event.Job == "" {
		http.Error(w, "job name is required", http.StatusBadRequest)
		return
	}
	ts := event.Timestamp
	if ts.IsZero() {
		ts = time.Now().UTC()
	}
	s.monitor.RecordRun(event.Job, ts)
	if event.Status == "failure" {
		s.monitor.MarkFailed(event.Job)
	}
	writeJSON(w, http.StatusAccepted, map[string]string{"accepted": event.Job})
}
