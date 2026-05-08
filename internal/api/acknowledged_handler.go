package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yourorg/cronwatch/internal/monitor"
)

type acknowledgedHandler struct {
	store *monitor.AcknowledgementStore
}

func newAcknowledgedHandler(store *monitor.AcknowledgementStore) *acknowledgedHandler {
	return &acknowledgedHandler{store: store}
}

// handleAcknowledge accepts POST /ack with JSON body {job, acked_by, note, duration}.
func (h *acknowledgedHandler) handleAcknowledge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job      string `json:"job"`
		AckedBy  string `json:"acked_by"`
		Note     string `json:"note"`
		Duration string `json:"duration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Job == "" || req.AckedBy == "" || req.Duration == "" {
		http.Error(w, "job, acked_by, and duration are required", http.StatusBadRequest)
		return
	}
	d, err := time.ParseDuration(req.Duration)
	if err != nil {
		http.Error(w, "invalid duration", http.StatusBadRequest)
		return
	}
	h.store.Acknowledge(req.Job, req.AckedBy, req.Note, d)
	w.WriteHeader(http.StatusNoContent)
}

// handleGetAck returns GET /ack?job=<name> acknowledgement status.
func (h *acknowledgedHandler) handleGetAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "job parameter required", http.StatusBadRequest)
		return
	}
	e, ok := h.store.Get(job)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	writeJSON(w, e)
}

// handleLiftAck accepts DELETE /ack?job=<name>.
func (h *acknowledgedHandler) handleLiftAck(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "job parameter required", http.StatusBadRequest)
		return
	}
	h.store.Lift(job)
	w.WriteHeader(http.StatusNoContent)
}
