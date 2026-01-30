package metaagent

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Handler exposes meta-agent operations over HTTP.
type Handler struct {
	agent *Agent
}

// NewHandler creates a new meta-agent HTTP handler.
func NewHandler(a *Agent) *Handler {
	return &Handler{agent: a}
}

// FanOut handles POST /api/v1/meta/fanout.
func (h *Handler) FanOut(w http.ResponseWriter, r *http.Request) {
	var req FanOutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Prompt == "" {
		writeError(w, http.StatusBadRequest, "prompt is required")
		return
	}

	resp, err := h.agent.FanOut(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "fan-out failed")
		slog.Error("fan-out failed", "error", err)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
