package api

import (
	"net/http"

	"github.com/example/cronwatch/internal/monitor"
)

// metricsProvider is satisfied by *monitor.MetricsStore.
type metricsProvider interface {
	All() []monitor.JobMetrics
	Get(jobName string) (monitor.JobMetrics, bool)
}

// handleMetrics returns aggregated metrics for all tracked jobs.
//
//	GET /metrics
func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, s.metrics.All())
}

// handleJobMetrics returns metrics for a single job identified by the
// "job" query parameter.
//
//	GET /metrics/job?job=<name>
func (s *Server) handleJobMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := r.URL.Query().Get("job")
	if name == "" {
		http.Error(w, "missing job query parameter", http.StatusBadRequest)
		return
	}

	met, ok := s.metrics.Get(name)
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	writeJSON(w, met)
}
