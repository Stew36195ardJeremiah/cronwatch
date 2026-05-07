package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronwatch/internal/monitor"
)

type pauseHandler struct {
	store *monitor.PauseStore
}

func newPauseHandler(store *monitor.PauseStore) *pauseHandler {
	return &pauseHandler{store: store}
}

func (h *pauseHandler) handlePause(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job      string `json:"job"`
		Reason   string `json:"reason"`
		PausedBy string `json:"paused_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Job == "" {
		http.Error(w, "missing job", http.StatusBadRequest)
		return
	}
	h.store.Pause(req.Job, req.Reason, req.PausedBy)
	writeJSON(w, http.StatusOK, map[string]string{"status": "paused", "job": req.Job})
}

func (h *pauseHandler) handleResume(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job string `json:"job"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Job == "" {
		http.Error(w, "missing job", http.StatusBadRequest)
		return
	}
	h.store.Resume(req.Job)
	writeJSON(w, http.StatusOK, map[string]string{"status": "resumed", "job": req.Job})
}

func (h *pauseHandler) handleListPaused(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, h.store.All())
}
