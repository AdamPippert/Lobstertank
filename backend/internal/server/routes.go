package server

import (
	"net/http"

	"github.com/AdamPippert/Lobstertank/internal/auth"
	"github.com/AdamPippert/Lobstertank/internal/gateway"
	"github.com/AdamPippert/Lobstertank/internal/metaagent"
)

func registerRoutes(
	mux *http.ServeMux,
	gw *gateway.Handler,
	meta *metaagent.Handler,
	authProvider auth.Provider,
) {
	authMW := auth.Middleware(authProvider)

	// Health check — unauthenticated.
	mux.HandleFunc("GET /healthz", handleHealthz)

	// Gateway CRUD — authenticated.
	mux.Handle("GET /api/v1/gateways", authMW(http.HandlerFunc(gw.List)))
	mux.Handle("POST /api/v1/gateways", authMW(http.HandlerFunc(gw.Create)))
	mux.Handle("GET /api/v1/gateways/{id}", authMW(http.HandlerFunc(gw.Get)))
	mux.Handle("PUT /api/v1/gateways/{id}", authMW(http.HandlerFunc(gw.Update)))
	mux.Handle("DELETE /api/v1/gateways/{id}", authMW(http.HandlerFunc(gw.Delete)))

	// Gateway actions.
	mux.Handle("POST /api/v1/gateways/{id}/health", authMW(http.HandlerFunc(gw.HealthCheck)))

	// Meta-agent — fan-out.
	mux.Handle("POST /api/v1/meta/fanout", authMW(http.HandlerFunc(meta.FanOut)))
}

func handleHealthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
