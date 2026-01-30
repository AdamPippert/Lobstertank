package gateway

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/AdamPippert/Lobstertank/internal/audit"
	"github.com/AdamPippert/Lobstertank/internal/model"
)

// Handler exposes gateway CRUD operations over HTTP.
type Handler struct {
	registry      *Registry
	clientFactory *ClientFactory
	auditor       *audit.Logger
}

// NewHandler constructs a gateway HTTP handler.
func NewHandler(r *Registry, cf *ClientFactory, a *audit.Logger) *Handler {
	return &Handler{registry: r, clientFactory: cf, auditor: a}
}

// List handles GET /api/v1/gateways.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	gateways, err := h.registry.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list gateways", err)
		return
	}
	writeJSON(w, http.StatusOK, gateways)
}

// Create handles POST /api/v1/gateways.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateGatewayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	if req.Name == "" || req.Endpoint == "" {
		writeError(w, http.StatusBadRequest, "name and endpoint are required", nil)
		return
	}

	gw, err := h.registry.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create gateway", err)
		return
	}

	writeJSON(w, http.StatusCreated, gw)
}

// Get handles GET /api/v1/gateways/{id}.
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	gw, err := h.registry.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "gateway not found", err)
		return
	}
	writeJSON(w, http.StatusOK, gw)
}

// Update handles PUT /api/v1/gateways/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req model.UpdateGatewayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err)
		return
	}

	gw, err := h.registry.Update(r.Context(), id, req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update gateway", err)
		return
	}

	writeJSON(w, http.StatusOK, gw)
}

// Delete handles DELETE /api/v1/gateways/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.registry.Delete(r.Context(), id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete gateway", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// HealthCheck handles POST /api/v1/gateways/{id}/health.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	gw, err := h.registry.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "gateway not found", err)
		return
	}

	client := h.clientFactory.ClientFor(gw)
	result, _ := client.HealthCheck(r.Context())

	// Update the stored status regardless of probe outcome.
	if err := h.registry.UpdateStatus(r.Context(), id, result.Status); err != nil {
		slog.Warn("failed to persist gateway status", "id", id, "error", err)
	}

	writeJSON(w, http.StatusOK, result)
}

// --- helpers ---

type apiError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string, err error) {
	resp := apiError{Error: msg}
	if err != nil {
		slog.Error(msg, "error", err)
	}
	writeJSON(w, status, resp)
}
