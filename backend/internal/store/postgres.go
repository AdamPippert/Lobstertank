package store

import (
	"context"
	"fmt"
	"time"

	"github.com/AdamPippert/Lobstertank/internal/model"
)

// PostgresStore implements Store using PostgreSQL.
// TODO: Implement with database/sql + github.com/lib/pq or pgx.
type PostgresStore struct {
	dsn string
}

// NewPostgresStore creates a PostgreSQL-backed store.
func NewPostgresStore(dsn string) (*PostgresStore, error) {
	// TODO: Establish connection pool, run migrations.
	return &PostgresStore{dsn: dsn}, nil
}

func (s *PostgresStore) ListGateways(_ context.Context) ([]model.Gateway, error) {
	return nil, fmt.Errorf("PostgresStore.ListGateways not yet implemented")
}

func (s *PostgresStore) GetGateway(_ context.Context, id string) (*model.Gateway, error) {
	return nil, fmt.Errorf("PostgresStore.GetGateway(%s) not yet implemented", id)
}

func (s *PostgresStore) CreateGateway(_ context.Context, _ *model.Gateway) error {
	return fmt.Errorf("PostgresStore.CreateGateway not yet implemented")
}

func (s *PostgresStore) UpdateGateway(_ context.Context, _ *model.Gateway) error {
	return fmt.Errorf("PostgresStore.UpdateGateway not yet implemented")
}

func (s *PostgresStore) DeleteGateway(_ context.Context, id string) error {
	return fmt.Errorf("PostgresStore.DeleteGateway(%s) not yet implemented", id)
}

func (s *PostgresStore) UpdateGatewayStatus(_ context.Context, id string, _ string, _ *time.Time) error {
	return fmt.Errorf("PostgresStore.UpdateGatewayStatus(%s) not yet implemented", id)
}

func (s *PostgresStore) Close() error {
	// TODO: Close connection pool.
	return nil
}
