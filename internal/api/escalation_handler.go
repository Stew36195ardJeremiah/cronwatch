package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/cronwatch/internal/monitor"
)

type escalationHandler struct {
	store *monitor.EscalationStore
}

func newEscalationHandler(s *monitor.EscalationStore) *escalationHandler {
	return &escalationHandler{store: s}
}

// handleSetPolicy registers an escalation policy for a job.
// POST /escalation/policy
func (h *escalationHandler) handleSetPolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job           string `json:"job"`
		WarnAfter     string `json:"warn_after"`
		CriticalAfter string `json:"critical_after"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Job == "" {
		http.Error(w, "missing job", http.StatusBadRequest)
		return
	}
	p := monitor.EscalationPolicy{JobName: req.Job}
	if req.WarnAfter != "" {
		d, err := time.ParseDuration(req.WarnAfter)
		if err != nil {
			http.Error(w, "invalid warn_after duration", http.StatusBadRequest)
			return
		}
		p.WarnAfter = d
	}
	if req.CriticalAfter != "" {
		d, err := time.ParseDuration(req.CriticalAfter)
		if err != nil {
			http.Error(w, "invalid critical_after duration", http.StatusBadRequest)
			return
		}
		p.CriticalAfter = d
	}
	h.store.SetPolicy(p)
	w.WriteHeader(http.StatusNoContent)
}

// handleStatus returns current escalation levels for all triggered jobs.
// GET /escalation/status
func (h *escalationHandler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	all := h.store.All()
	out := make(map[string]string, len(all))
	for job, lvl := range all {
		out[job] = lvl.String()
	}
	writeJSON(w, out)
}

// handleReset clears the escalation state for a job.
// POST /escalation/reset
func (h *escalationHandler) handleReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job param", http.StatusBadRequest)
		return
	}
	h.store.Reset(job)
	w.WriteHeader(http.StatusNoContent)
}
