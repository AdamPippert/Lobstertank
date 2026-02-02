package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/AdamPippert/Lobstertank/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// PostgresStore implements Store using PostgreSQL via pgx.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a PostgreSQL-backed store.
// The dsn should be a PostgreSQL connection string, e.g.,
// "postgres://user:pass@localhost:5432/lobstertank?sslmode=disable".
func NewPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}

	// Configure connection pool.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Verify connectivity.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	// Run migrations.
	if _, err := db.ExecContext(ctx, createGatewaysTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("create gateways table: %w", err)
	}

	slog.Info("postgres store initialized")
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) ListGateways(ctx context.Context) ([]model.Gateway, error) {
	query := fmt.Sprintf("SELECT %s FROM gateways ORDER BY enrolled_at DESC", gatewayColumns)
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query gateways: %w", err)
	}
	defer rows.Close()

	gateways := make([]model.Gateway, 0)
	for rows.Next() {
		gw, err := scanGateway(rows)
		if err != nil {
			return nil, fmt.Errorf("scan gateway row: %w", err)
		}
		gateways = append(gateways, *gw)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate gateway rows: %w", err)
	}

	return gateways, nil
}

func (s *PostgresStore) GetGateway(ctx context.Context, id string) (*model.Gateway, error) {
	query := fmt.Sprintf("SELECT %s FROM gateways WHERE id = $1", gatewayColumns)
	row := s.db.QueryRowContext(ctx, query, id)
	gw, err := scanGateway(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("gateway not found: %s", id)
		}
		return nil, fmt.Errorf("scan gateway: %w", err)
	}
	return gw, nil
}

func (s *PostgresStore) CreateGateway(ctx context.Context, gw *model.Gateway) error {
	query := `INSERT INTO gateways (
        id, name, description, endpoint,
        transport_type, transport_params,
        auth_type, auth_params, auth_secret_ref,
        status, labels, enrolled_at, last_seen_at, ttl_seconds
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

	var ttl *int64
	if gw.TTLSeconds != nil {
		v := int64(*gw.TTLSeconds)
		ttl = &v
	}

	_, err := s.db.ExecContext(ctx, query,
		gw.ID,
		gw.Name,
		gw.Description,
		gw.Endpoint,
		gw.Transport.Type,
		marshalJSONMap(gw.Transport.Params),
		gw.Auth.Type,
		marshalJSONMap(gw.Auth.Params),
		gw.Auth.SecretRef,
		string(gw.Status),
		marshalJSONMap(gw.Labels),
		gw.EnrolledAt,
		gw.LastSeenAt,
		ttl,
	)
	if err != nil {
		return fmt.Errorf("insert gateway: %w", err)
	}
	return nil
}

func (s *PostgresStore) UpdateGateway(ctx context.Context, gw *model.Gateway) error {
	query := `UPDATE gateways SET
        name = $2,
        description = $3,
        endpoint = $4,
        transport_type = $5,
        transport_params = $6,
        auth_type = $7,
        auth_params = $8,
        auth_secret_ref = $9,
        status = $10,
        labels = $11,
        last_seen_at = $12,
        ttl_seconds = $13
    WHERE id = $1`

	var ttl *int64
	if gw.TTLSeconds != nil {
		v := int64(*gw.TTLSeconds)
		ttl = &v
	}

	result, err := s.db.ExecContext(ctx, query,
		gw.ID,
		gw.Name,
		gw.Description,
		gw.Endpoint,
		gw.Transport.Type,
		marshalJSONMap(gw.Transport.Params),
		gw.Auth.Type,
		marshalJSONMap(gw.Auth.Params),
		gw.Auth.SecretRef,
		string(gw.Status),
		marshalJSONMap(gw.Labels),
		gw.LastSeenAt,
		ttl,
	)
	if err != nil {
		return fmt.Errorf("update gateway: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("gateway not found: %s", gw.ID)
	}
	return nil
}

func (s *PostgresStore) DeleteGateway(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM gateways WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("delete gateway: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("gateway not found: %s", id)
	}
	return nil
}

func (s *PostgresStore) UpdateGatewayStatus(ctx context.Context, id string, status string, lastSeen *time.Time) error {
	result, err := s.db.ExecContext(ctx,
		"UPDATE gateways SET status = $2, last_seen_at = $3 WHERE id = $1",
		id, status, lastSeen,
	)
	if err != nil {
		return fmt.Errorf("update gateway status: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("gateway not found: %s", id)
	}
	return nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}
