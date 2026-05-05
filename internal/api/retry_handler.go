package api

import (
	"net/http"

	"github.com/user/cronwatch/internal/monitor"
)

// retryProvider abstracts retry state access for the handler.
type retryProvider interface {
	State(job string) (monitor.RetryState, bool)
	Reset(job string)
}

type retryResponse struct {
	Job       string `json:"job"`
	Attempts  int    `json:"attempts"`
	Exhausted bool   `json:"exhausted"`
}

// handleRetryStatus returns the current retry state for a job.
func handleRetryStatus(store retryProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, "missing job parameter", http.StatusBadRequest)
			return
		}
		state, ok := store.State(job)
		if !ok {
			http.Error(w, "job not found", http.StatusNotFound)
			return
		}
		writeJSON(w, retryResponse{
			Job:       job,
			Attempts:  state.Attempts,
			Exhausted: state.Exhausted,
		})
	}
}

// handleRetryReset clears the retry state for a job.
func handleRetryReset(store retryProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		job := r.URL.Query().Get("job")
		if job == "" {
			http.Error(w, "missing job parameter", http.StatusBadRequest)
			return
		}
		store.Reset(job)
		w.WriteHeader(http.StatusNoContent)
	}
}
