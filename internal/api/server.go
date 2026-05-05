package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cronwatch/internal/monitor"
)

// Server exposes a lightweight HTTP API for querying monitor state.
type Server struct {
	addr    string
	monitor *monitor.Monitor
	httpSrv *http.Server
}

// New creates a new API server bound to addr.
func New(addr string, m *monitor.Monitor) *Server {
	s := &Server{
		addr:    addr,
		monitor: m,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/history", s.handleHistory)

	s.httpSrv = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return s
}

// Start begins listening. It blocks until the server stops.
func (s *Server) Start() error {
	return s.httpSrv.ListenAndServe()
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() error {
	return s.httpSrv.Close()
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
