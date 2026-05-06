package api

import (
	"net/http"

	"github.com/user/cronwatch/internal/monitor"
)

type rateLimitHandler struct {
	store *monitor.RateLimitStore
}

func newRateLimitHandler(store *monitor.RateLimitStore) *rateLimitHandler {
	return &rateLimitHandler{store: store}
}

// handleRateLimitStatus returns the current rate limit snapshot for all jobs.
func (h *rateLimitHandler) handleRateLimitStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	all := h.store.All()
	type entry struct {
		Job         string `json:"job"`
		Count       int    `json:"count"`
		LastAlertAt string `json:"last_alert_at"`
	}
	result := make([]entry, 0, len(all))
	for job, e := range all {
		result = append(result, entry{
			Job:         job,
			Count:       e.Count,
			LastAlertAt: e.LastAlertAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	writeJSON(w, http.StatusOK, result)
}

// handleRateLimitReset resets the rate limit counter for a specific job.
func (h *rateLimitHandler) handleRateLimitReset(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, map[string]string{"status": "reset", "job": job})
}
