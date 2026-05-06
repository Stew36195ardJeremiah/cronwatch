package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/user/cronwatch/internal/monitor"
)

type dependencyHandler struct {
	store *monitor.DependencyStore
}

func newDependencyHandler(s *monitor.DependencyStore) *dependencyHandler {
	return &dependencyHandler{store: s}
}

// handleAddEdge POST /api/dependencies  {"from":"job-a","to":"job-b"}
func (h *dependencyHandler) handleAddEdge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		From string `json:"from"`
		To   string `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.From == "" || body.To == "" {
		http.Error(w, "invalid request body; 'from' and 'to' required", http.StatusBadRequest)
		return
	}
	h.store.AddEdge(body.From, body.To)
	w.WriteHeader(http.StatusCreated)
	writeJSON(w, map[string]string{"status": "added", "from": body.From, "to": body.To})
}

// handleListEdges GET /api/dependencies
func (h *dependencyHandler) handleListEdges(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, h.store.Edges())
}

// handleBlocked GET /api/dependencies/blocked?job=<name>&since=<duration>
func (h *dependencyHandler) handleBlocked(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing 'job' query param", http.StatusBadRequest)
		return
	}
	sinceStr := r.URL.Query().Get("since")
	sinceD, err := time.ParseDuration(sinceStr)
	if err != nil {
		sinceD = time.Hour
	}
	since := time.Now().Add(-sinceD)
	blockErr := h.store.Blocked(job, since)
	payload := map[string]interface{}{"job": job, "blocked": blockErr != nil}
	if blockErr != nil {
		payload["reason"] = blockErr.Error()
	}
	writeJSON(w, payload)
}
