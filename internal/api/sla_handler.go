package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronwatch/internal/monitor"
)

type slaHandler struct {
	store *monitor.SLAStore
}

func newSLAHandler(store *monitor.SLAStore) *slaHandler {
	return &slaHandler{store: store}
}

// handleSLASet registers or updates an SLA policy for a job.
// POST /sla/set  {"job":"...","max_duration":"5m","deadline":"2024-01-01T06:00:00Z"}
func (h *slaHandler) handleSLASet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job         string `json:"job"`
		MaxDuration string `json:"max_duration"`
		Deadline    string `json:"deadline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Job == "" {
		http.Error(w, "job is required", http.StatusBadRequest)
		return
	}
	var maxDur time.Duration
	if req.MaxDuration != "" {
		var err error
		maxDur, err = time.ParseDuration(req.MaxDuration)
		if err != nil {
			http.Error(w, "invalid max_duration", http.StatusBadRequest)
			return
		}
	}
	var deadline time.Time
	if req.Deadline != "" {
		var err error
		deadline, err = time.Parse(time.RFC3339, req.Deadline)
		if err != nil {
			http.Error(w, "invalid deadline (use RFC3339)", http.StatusBadRequest)
			return
		}
	}
	h.store.Set(req.Job, maxDur, deadline)
	w.WriteHeader(http.StatusNoContent)
}

// handleSLAStatus returns all SLA entries or a single job's entry.
// GET /sla/status?job=<name>
func (h *slaHandler) handleSLAStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job != "" {
		e, ok := h.store.Get(job)
		if !ok {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		writeJSON(w, e)
		return
	}
	writeJSON(w, h.store.All())
}
