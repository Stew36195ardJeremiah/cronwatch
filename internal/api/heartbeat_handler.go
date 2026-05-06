package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/yourusername/cronwatch/internal/monitor"
)

type heartbeatHandler struct {
	store *monitor.HeartbeatStore
}

func newHeartbeatHandler(store *monitor.HeartbeatStore) *heartbeatHandler {
	return &heartbeatHandler{store: store}
}

// handleBeat accepts a POST with ?job=<name>&ttl=<seconds> and records a heartbeat.
func (h *heartbeatHandler) handleBeat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	ttlStr := r.URL.Query().Get("ttl")
	if ttlStr == "" {
		http.Error(w, "missing ttl parameter", http.StatusBadRequest)
		return
	}
	ttlSecs, err := strconv.ParseFloat(ttlStr, 64)
	if err != nil || ttlSecs <= 0 {
		http.Error(w, "invalid ttl: must be a positive number of seconds", http.StatusBadRequest)
		return
	}
	h.store.Beat(job, time.Duration(ttlSecs*float64(time.Second)))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "job": job})
}

// handleHeartbeatStatus returns all heartbeat records with their current expiry state.
func (h *heartbeatHandler) handleHeartbeatStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	all := h.store.All()
	type row struct {
		Job      string  `json:"job"`
		LastSeen string  `json:"last_seen"`
		TTLSecs  float64 `json:"ttl_seconds"`
		Expired  bool    `json:"expired"`
	}
	result := make([]row, 0, len(all))
	for _, rec := range all {
		result = append(result, row{
			Job:      rec.JobName,
			LastSeen: rec.LastSeen.UTC().Format(time.RFC3339),
			TTLSecs:  rec.TTL.Seconds(),
			Expired:  rec.Expired,
		})
	}
	writeJSON(w, http.StatusOK, result)
}
