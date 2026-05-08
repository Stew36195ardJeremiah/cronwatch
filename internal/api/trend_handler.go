package api

import (
	"net/http"

	"github.com/cronwatch/cronwatch/internal/monitor"
)

type trendHandler struct {
	store *monitor.TrendStore
}

func newTrendHandler(s *monitor.TrendStore) *trendHandler {
	return &trendHandler{store: s}
}

// handleTrendAll returns trend statistics for all tracked jobs.
func (h *trendHandler) handleTrendAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	entries := h.store.All()
	if entries == nil {
		entries = []monitor.TrendEntry{}
	}
	writeJSON(w, entries)
}

// handleTrendGet returns trend statistics for a single job.
func (h *trendHandler) handleTrendGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	entry, ok := h.store.Get(job)
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	writeJSON(w, entry)
}
