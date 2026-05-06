package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/cronwatch/internal/monitor"
)

type tagHandler struct {
	store *monitor.TagStore
}

func newTagHandler(store *monitor.TagStore) *tagHandler {
	return &tagHandler{store: store}
}

// handleGetTags returns all tags for a job: GET /tags?job=<name>
func (h *tagHandler) handleGetTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job param", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"job":  job,
		"tags": h.store.Get(job),
	})
}

// handleSetTags replaces tags for a job: POST /tags
func (h *tagHandler) handleSetTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job  string   `json:"job"`
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	h.store.Set(req.Job, req.Tags)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// handleJobsWithTag returns jobs carrying a tag: GET /tags/jobs?tag=<name>
func (h *tagHandler) handleJobsWithTag(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	tag := r.URL.Query().Get("tag")
	if tag == "" {
		http.Error(w, "missing tag param", http.StatusBadRequest)
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tag":  tag,
		"jobs": h.store.JobsWithTag(tag),
	})
}
