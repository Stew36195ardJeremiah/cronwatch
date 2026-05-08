package api

import (
	"net/http"

	"github.com/cronwatch/internal/monitor"
)

type flapHandler struct {
	store *monitor.FlapStore
}

func newFlapHandler(store *monitor.FlapStore) *flapHandler {
	return &flapHandler{store: store}
}

// handleFlapStatus returns the flap state for a specific job.
//
// GET /api/flap/status?job=<name>
func (h *flapHandler) handleFlapStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	entry := h.store.Get(job)
	if entry == nil {
		writeJSON(w, map[string]interface{}{
			"job":         job,
			"is_flapping": false,
			"transitions": 0,
		})
		return
	}
	writeJSON(w, map[string]interface{}{
		"job":          job,
		"is_flapping":  entry.IsFlapping,
		"transitions":  len(entry.Transitions),
		"last_checked": entry.LastChecked,
	})
}

// handleFlapAll returns flap state for all tracked jobs.
//
// GET /api/flap/all
func (h *flapHandler) handleFlapAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	all := h.store.All()
	type row struct {
		Job         string `json:"job"`
		IsFlapping  bool   `json:"is_flapping"`
		Transitions int    `json:"transitions"`
	}
	rows := make([]row, 0, len(all))
	for job, e := range all {
		rows = append(rows, row{
			Job:        job,
			IsFlapping: e.IsFlapping,
			Transitions: len(e.Transitions),
		})
	}
	writeJSON(w, rows)
}

// handleFlapReset clears flap state for a job.
//
// POST /api/flap/reset?job=<name>
func (h *flapHandler) handleFlapReset(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	h.store.Reset(job)
	writeJSON(w, map[string]string{"status": "ok", "job": job})
}
