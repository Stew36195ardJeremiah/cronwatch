package api

import (
	"net/http"
	"time"
)

type healthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

type statusEntry struct {
	Job       string     `json:"job"`
	LastRun   *time.Time `json:"last_run,omitempty"`
	NextRun   *time.Time `json:"next_run,omitempty"`
	Overdue   bool       `json:"overdue"`
	DriftSecs float64    `json:"drift_seconds"`
	Failed    bool       `json:"failed"`
}

type historyEntry struct {
	Job       string    `json:"job"`
	Timestamp time.Time `json:"timestamp"`
	Success   bool      `json:"success"`
	DriftSecs float64   `json:"drift_seconds"`
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status: "ok",
		Time:   time.Now().UTC(),
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	all := s.monitor.Statuses()
	result := make([]statusEntry, 0, len(all))
	for name, st := range all {
		e := statusEntry{
			Job:     name,
			Overdue: st.Overdue,
			Failed:  st.Failed,
		}
		if !st.LastRun.IsZero() {
			t := st.LastRun
			e.LastRun = &t
		}
		if !st.NextRun.IsZero() {
			t := st.NextRun
			e.NextRun = &t
		}
		e.DriftSecs = st.Drift.Seconds()
		result = append(result, e)
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, `{"error":"missing job query param"}`, http.StatusBadRequest)
		return
	}
	records := s.monitor.History(job)
	result := make([]historyEntry, 0, len(records))
	for _, rec := range records {
		result = append(result, historyEntry{
			Job:       job,
			Timestamp: rec.Timestamp,
			Success:   rec.Success,
			DriftSecs: rec.Drift.Seconds(),
		})
	}
	writeJSON(w, http.StatusOK, result)
}
