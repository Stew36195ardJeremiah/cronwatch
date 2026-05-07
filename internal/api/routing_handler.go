package api

import (
	"encoding/json"
	"net/http"

	"github.com/cronwatch/cronwatch/internal/monitor"
)

type routingHandler struct {
	store *monitor.RoutingStore
}

func newRoutingHandler(store *monitor.RoutingStore) *routingHandler {
	return &routingHandler{store: store}
}

// handleSetRoute handles POST /routes — sets a job->channel mapping.
func (h *routingHandler) handleSetRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Job     string `json:"job"`
		Channel string `json:"channel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if err := h.store.Set(req.Job, req.Channel); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleDeleteRoute handles DELETE /routes?job=<name>.
func (h *routingHandler) handleDeleteRoute(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	job := r.URL.Query().Get("job")
	if job == "" {
		http.Error(w, "missing job param", http.StatusBadRequest)
		return
	}
	h.store.Remove(job)
	w.WriteHeader(http.StatusNoContent)
}

// handleListRoutes handles GET /routes.
func (h *routingHandler) handleListRoutes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	type response struct {
		Default string               `json:"default"`
		Routes  []monitor.RouteRule  `json:"routes"`
	}
	writeJSON(w, response{
		Default: h.store.Default(),
		Routes:  h.store.All(),
	})
}
