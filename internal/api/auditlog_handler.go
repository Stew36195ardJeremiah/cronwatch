package api

import (
	"net/http"

	"github.com/example/cronwatch/internal/monitor"
)

type auditLogHandler struct {
	store *monitor.AuditLogStore
}

func newAuditLogHandler(store *monitor.AuditLogStore) *auditLogHandler {
	return &auditLogHandler{store: store}
}

// handleAuditAll returns all audit log entries.
func (h *auditLogHandler) handleAuditAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, h.store.All())
}

// handleAuditForJob returns audit log entries filtered to a single job.
func (h *auditLogHandler) handleAuditForJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job parameter", http.StatusBadRequest)
		return
	}
	writeJSON(w, h.store.ForJob(job))
}

// RegisterAuditRoutes attaches audit log endpoints to the given mux.
func RegisterAuditRoutes(mux *http.ServeMux, store *monitor.AuditLogStore) {
	h := newAuditLogHandler(store)
	mux.HandleFunc("/audit", h.handleAuditAll)
	mux.HandleFunc("/audit/job", h.handleAuditForJob)
}
