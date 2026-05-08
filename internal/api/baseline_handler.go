package api

import (
	"net/http"

	"github.com/example/cronwatch/internal/monitor"
)

type baselineHandler struct {
	store *monitor.BaselineStore
}

func newBaselineHandler(store *monitor.BaselineStore) *baselineHandler {
	return &baselineHandler{store: store}
}

// handleBaselineAll returns all baseline entries.
func (h *baselineHandler) handleBaselineAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, h.store.All())
}

// handleBaselineGet returns the baseline for a single job.
func (h *baselineHandler) handleBaselineGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job param", http.StatusBadRequest)
		return
	}
	e, ok := h.store.Get(job)
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	writeJSON(w, e)
}

// handleBaselineReset clears the baseline for a job.
func (h *baselineHandler) handleBaselineReset(w http.ResponseWriter, r *http.Request) {
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
