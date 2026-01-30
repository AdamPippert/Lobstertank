package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/AdamPippert/Lobstertank/internal/audit"
	"github.com/AdamPippert/Lobstertank/internal/auth"
	"github.com/AdamPippert/Lobstertank/internal/config"
	"github.com/AdamPippert/Lobstertank/internal/gateway"
	"github.com/AdamPippert/Lobstertank/internal/metaagent"
)

// Dependencies holds all injected service dependencies for the server.
type Dependencies struct {
	Config        *config.Config
	Registry      *gateway.Registry
	ClientFactory *gateway.ClientFactory
	MetaAgent     *metaagent.Agent
	AuthProvider  auth.Provider
	Auditor       *audit.Logger
}

// Server wraps the net/http.Server with application-specific setup.
type Server struct {
	httpServer *http.Server
	deps       Dependencies
}

// New creates a configured Server ready to run.
func New(deps Dependencies) *Server {
	mux := http.NewServeMux()

	gatewayHandler := gateway.NewHandler(deps.Registry, deps.ClientFactory, deps.Auditor)
	metaHandler := metaagent.NewHandler(deps.MetaAgent)

	registerRoutes(mux, gatewayHandler, metaHandler, deps.AuthProvider)

	addr := fmt.Sprintf("%s:%d", deps.Config.Server.Host, deps.Config.Server.Port)

	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		deps: deps,
	}
}

// Run starts the HTTP server and blocks until the context is canceled.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		slog.Info("lobstertank server starting", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		slog.Info("shutdown signal received, draining connections")
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	slog.Info("server stopped gracefully")
	return nil
}
