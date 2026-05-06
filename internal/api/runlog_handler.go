package api

import (
	"net/http"
	"strconv"

	"github.com/user/cronwatch/internal/monitor"
)

type runLogHandler struct {
	log *monitor.RunLog
}

func newRunLogHandler(log *monitor.RunLog) http.Handler {
	return &runLogHandler{log: log}
}

func (h *runLogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing required query param: job", http.StatusBadRequest)
		return
	}

	entries := h.log.ForJob(job)

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 && n < len(entries) {
			entries = entries[len(entries)-n:]
		}
	}

	if entries == nil {
		entries = []monitor.RunEntry{}
	}

	writeJSON(w, http.StatusOK, entries)
}
