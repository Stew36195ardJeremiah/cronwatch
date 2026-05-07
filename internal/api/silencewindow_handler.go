package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/example/cronwatch/internal/monitor"
)

type silenceWindowHandler struct {
	store *monitor.SilenceWindowStore
}

func newSilenceWindowHandler(store *monitor.SilenceWindowStore) *silenceWindowHandler {
	return &silenceWindowHandler{store: store}
}

// handleAdd registers a new silence window via POST /silence.
func (h *silenceWindowHandler) handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job      string `json:"job"`
		Start    string `json:"start"`
		End      string `json:"end"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Job == "" || req.Start == "" || req.End == "" {
		http.Error(w, "job, start and end are required", http.StatusBadRequest)
		return
	}
	start, err := time.Parse(time.RFC3339, req.Start)
	if err != nil {
		http.Error(w, "invalid start time: use RFC3339", http.StatusBadRequest)
		return
	}
	end, err := time.Parse(time.RFC3339, req.End)
	if err != nil {
		http.Error(w, "invalid end time: use RFC3339", http.StatusBadRequest)
		return
	}
	h.store.Add(monitor.SilenceWindow{
		JobName: req.Job,
		Start:   start,
		End:     end,
		Reason:  req.Reason,
	})
	w.WriteHeader(http.StatusCreated)
}

// handleList returns all current silence windows via GET /silence.
func (h *silenceWindowHandler) handleList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, h.store.All())
}

// handleCheck reports whether a specific job is currently silenced.
func (h *silenceWindowHandler) handleCheck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "job parameter required", http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]bool{"silenced": h.store.IsSilenced(job)})
}
