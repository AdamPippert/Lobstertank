package store

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/AdamPippert/Lobstertank/internal/model"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStore implements Store using SQLite via mattn/go-sqlite3.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a SQLite-backed store.
// The dsn is the database file path (e.g., "lobstertank.db") or ":memory:"
// for an in-memory database.
func NewSQLiteStore(dsn string) (*SQLiteStore, error) {
	if dsn == "" {
		dsn = ":memory:"
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("open sqlite connection: %w", err)
	}

	// SQLite performs best with a single writer connection.
	db.SetMaxOpenConns(1)

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}

	// Enable foreign keys.
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	// Run migrations.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, createGatewaysTableSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("create gateways table: %w", err)
	}

	slog.Info("sqlite store initialized", "dsn", dsn)
	return &SQLiteStore{db: db}, nil
}

func (s *SQLiteStore) ListGateways(ctx context.Context) ([]model.Gateway, error) {
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

func (s *SQLiteStore) GetGateway(ctx context.Context, id string) (*model.Gateway, error) {
	query := fmt.Sprintf("SELECT %s FROM gateways WHERE id = ?", gatewayColumns)
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

func (s *SQLiteStore) CreateGateway(ctx context.Context, gw *model.Gateway) error {
	query := `INSERT INTO gateways (
        id, name, description, endpoint,
        transport_type, transport_params,
        auth_type, auth_params, auth_secret_ref,
        status, labels, enrolled_at, last_seen_at, ttl_seconds
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

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
		gw.EnrolledAt.Format(time.RFC3339),
		gw.LastSeenAt,
		ttl,
	)
	if err != nil {
		return fmt.Errorf("insert gateway: %w", err)
	}
	return nil
}

func (s *SQLiteStore) UpdateGateway(ctx context.Context, gw *model.Gateway) error {
	query := `UPDATE gateways SET
        name = ?,
        description = ?,
        endpoint = ?,
        transport_type = ?,
        transport_params = ?,
        auth_type = ?,
        auth_params = ?,
        auth_secret_ref = ?,
        status = ?,
        labels = ?,
        last_seen_at = ?,
        ttl_seconds = ?
    WHERE id = ?`

	var ttl *int64
	if gw.TTLSeconds != nil {
		v := int64(*gw.TTLSeconds)
		ttl = &v
	}

	result, err := s.db.ExecContext(ctx, query,
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
		gw.ID,
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

func (s *SQLiteStore) DeleteGateway(ctx context.Context, id string) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM gateways WHERE id = ?", id)
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

func (s *SQLiteStore) UpdateGatewayStatus(ctx context.Context, id string, status string, lastSeen *time.Time) error {
	result, err := s.db.ExecContext(ctx,
		"UPDATE gateways SET status = ?, last_seen_at = ? WHERE id = ?",
		status, lastSeen, id,
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

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}
