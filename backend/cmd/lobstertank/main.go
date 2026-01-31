package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/AdamPippert/Lobstertank/internal/audit"
	"github.com/AdamPippert/Lobstertank/internal/auth"
	"github.com/AdamPippert/Lobstertank/internal/cli"
	"github.com/AdamPippert/Lobstertank/internal/config"
	"github.com/AdamPippert/Lobstertank/internal/gateway"
	"github.com/AdamPippert/Lobstertank/internal/metaagent"
	"github.com/AdamPippert/Lobstertank/internal/secrets"
	"github.com/AdamPippert/Lobstertank/internal/server"
	"github.com/AdamPippert/Lobstertank/internal/store"
	"github.com/AdamPippert/Lobstertank/internal/transport"
)

func main() {
	// Dispatch CLI subcommands. If the command is "serve" (or no args),
	// cli.Run returns -1 and we fall through to the server boot below.
	if code := cli.Run(os.Args); code >= 0 {
		os.Exit(code)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Initialize audit logger.
	auditor := audit.New(cfg.Audit)

	// Initialize secrets provider.
	secretProvider, err := secrets.NewProvider(cfg.Secrets)
	if err != nil {
		slog.Error("failed to initialize secrets provider", "error", err)
		os.Exit(1)
	}

	// Initialize data store.
	dataStore, err := store.New(cfg.Database)
	if err != nil {
		slog.Error("failed to initialize data store", "error", err)
		os.Exit(1)
	}
	defer dataStore.Close()

	// Initialize transport provider.
	transportProvider := transport.NewProvider(cfg.Transport)

	// Initialize auth provider.
	authProvider, err := auth.NewProvider(cfg.Auth, secretProvider)
	if err != nil {
		slog.Error("failed to initialize auth provider", "error", err)
		os.Exit(1)
	}

	// Initialize gateway registry.
	registry := gateway.NewRegistry(dataStore, auditor)

	// Initialize gateway client factory.
	clientFactory := gateway.NewClientFactory(transportProvider, secretProvider)

	// Initialize meta-agent.
	agent := metaagent.New(registry, clientFactory, auditor)

	// Build and start the HTTP server.
	srv := server.New(server.Dependencies{
		Config:        cfg,
		Registry:      registry,
		ClientFactory: clientFactory,
		MetaAgent:     agent,
		AuthProvider:  authProvider,
		Auditor:       auditor,
	})

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := srv.Run(ctx); err != nil {
		slog.Error("server exited with error", "error", err)
		os.Exit(1)
	}

	slog.Info("lobstertank shutdown complete")
}
