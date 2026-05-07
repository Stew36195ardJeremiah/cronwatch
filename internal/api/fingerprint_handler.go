package api

import (
	"net/http"

	"github.com/yourorg/cronwatch/internal/monitor"
)

type fingerprintHandler struct {
	store *monitor.FingerprintStore
}

func newFingerprintHandler(store *monitor.FingerprintStore) *fingerprintHandler {
	return &fingerprintHandler{store: store}
}

// handleListFingerprints returns all active deduplicated alert fingerprints.
func (h *fingerprintHandler) handleListFingerprints(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	entries := h.store.All()
	type response struct {
		Hash      string `json:"hash"`
		JobName   string `json:"job"`
		Level     string `json:"level"`
		Message   string `json:"message"`
		Count     int    `json:"count"`
		FirstSeen string `json:"first_seen"`
		LastSeen  string `json:"last_seen"`
	}
	out := make([]response, 0, len(entries))
	for _, e := range entries {
		out = append(out, response{
			Hash:      e.Hash,
			JobName:   e.JobName,
			Level:     e.Level,
			Message:   e.Message,
			Count:     e.Count,
			FirstSeen: e.FirstSeen.UTC().Format("2006-01-02T15:04:05Z"),
			LastSeen:  e.LastSeen.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	writeJSON(w, http.StatusOK, out)
}

// handleRecordFingerprint manually records a fingerprint entry (useful for testing/CLI).
func (h *fingerprintHandler) handleRecordFingerprint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		Job     string `json:"job"`
		Level   string `json:"level"`
		Message string `json:"message"`
	}
	if err := readJSON(r, &body); err != nil || body.Job == "" || body.Level == "" {
		http.Error(w, "missing required fields: job, level", http.StatusBadRequest)
		return
	}
	isNew := h.store.Record(body.Job, body.Level, body.Message)
	writeJSON(w, http.StatusOK, map[string]bool{"new": isNew})
}
