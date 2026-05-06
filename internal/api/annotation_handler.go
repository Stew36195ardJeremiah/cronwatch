package api

import (
	"encoding/json"
	"net/http"

	"github.com/user/cronwatch/internal/monitor"
)

type annotationHandler struct {
	store *monitor.AnnotationStore
}

func newAnnotationHandler(store *monitor.AnnotationStore) *annotationHandler {
	return &annotationHandler{store: store}
}

// handleAddAnnotation handles POST /annotations?job=<name>
func (h *annotationHandler) handleAddAnnotation(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var body struct {
		JobName string `json:"job_name"`
		Author  string `json:"author"`
		Note    string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.JobName == "" || body.Note == "" {
		http.Error(w, "job_name and note are required", http.StatusBadRequest)
		return
	}
	h.store.Add(monitor.Annotation{
		JobName: body.JobName,
		Author:  body.Author,
		Note:    body.Note,
	})
	w.WriteHeader(http.StatusCreated)
}

// handleGetAnnotations handles GET /annotations?job=<name>
func (h *annotationHandler) handleGetAnnotations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "job parameter required", http.StatusBadRequest)
		return
	}
	list := h.store.Get(job)
	if list == nil {
		list = []monitor.Annotation{}
	}
	writeJSON(w, list)
}

// handleDeleteAnnotations handles DELETE /annotations?job=<name>
func (h *annotationHandler) handleDeleteAnnotations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "job parameter required", http.StatusBadRequest)
		return
	}
	h.store.Delete(job)
	w.WriteHeader(http.StatusNoContent)
}
