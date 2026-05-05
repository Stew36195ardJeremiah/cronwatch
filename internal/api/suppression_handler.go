package api

import (
	"encoding/json"
	"net/http"
	"time"
)

type suppressRequest struct {
	Job      string `json:"job"`
	Duration string `json:"duration"`
}

type suppressResponse struct {
	Job     string    `json:"job"`
	Until   time.Time `json:"until"`
	Message string    `json:"message"`
}

func (s *Server) handleSuppress(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req suppressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	dur, err := time.ParseDuration(req.Duration)
	if err != nil || dur <= 0 {
		http.Error(w, "invalid duration", http.StatusBadRequest)
		return
	}

	s.monitor.Suppress(req.Job, dur)

	writeJSON(w, http.StatusOK, suppressResponse{
		Job:     req.Job,
		Until:   time.Now().Add(dur),
		Message: "suppression applied",
	})
}

func (s *Server) handleLiftSuppression(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Job string `json:"job"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Job == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	s.monitor.LiftSuppression(req.Job)
	writeJSON(w, http.StatusOK, map[string]string{"job": req.Job, "message": "suppression lifted"})
}
