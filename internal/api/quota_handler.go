package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/yourorg/cronwatch/internal/monitor"
)

type quotaHandler struct {
	store *monitor.QuotaStore
}

func newQuotaHandler(store *monitor.QuotaStore) *quotaHandler {
	return &quotaHandler{store: store}
}

// handleQuotaStatus returns the current quota entry for a job.
func (h *quotaHandler) handleQuotaStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job param", http.StatusBadRequest)
		return
	}
	e := h.store.Get(job)
	if e == nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	writeJSON(w, e)
}

// handleQuotaSet sets a custom quota limit and window for a job.
func (h *quotaHandler) handleQuotaSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job    string `json:"job"`
		Limit  int    `json:"limit"`
		Window string `json:"window"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if req.Job == "" || req.Limit <= 0 || req.Window == "" {
		http.Error(w, "missing or invalid fields", http.StatusBadRequest)
		return
	}
	win, err := time.ParseDuration(req.Window)
	if err != nil {
		http.Error(w, "invalid window duration", http.StatusBadRequest)
		return
	}
	h.store.SetLimit(req.Job, req.Limit, win)
	w.WriteHeader(http.StatusNoContent)
}

// handleQuotaReset resets the counter for a job.
func (h *quotaHandler) handleQuotaReset(w http.ResponseWriter, r *http.Request) {
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
	_ = strconv.Itoa(0) // suppress unused import
}
