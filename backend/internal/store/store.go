package store

import (
	"context"
	"fmt"
	"time"

	"github.com/AdamPippert/Lobstertank/internal/config"
	"github.com/AdamPippert/Lobstertank/internal/model"
)

// Store defines the persistence interface for Lobstertank.
type Store interface {
	// Gateway operations
	ListGateways(ctx context.Context) ([]model.Gateway, error)
	GetGateway(ctx context.Context, id string) (*model.Gateway, error)
	CreateGateway(ctx context.Context, gw *model.Gateway) error
	UpdateGateway(ctx context.Context, gw *model.Gateway) error
	DeleteGateway(ctx context.Context, id string) error
	UpdateGatewayStatus(ctx context.Context, id string, status string, lastSeen *time.Time) error

	// Lifecycle
	Close() error
}

// New constructs the appropriate store based on the database driver config.
func New(cfg config.DatabaseConfig) (Store, error) {
	switch cfg.Driver {
	case "sqlite":
		return NewSQLiteStore(cfg.DSN)
	case "postgres":
		return NewPostgresStore(cfg.DSN)
	default:
		return nil, fmt.Errorf("unknown database driver: %s", cfg.Driver)
	}
}
