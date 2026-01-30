package store

import (
	"database/sql"
	"encoding/json"

	"github.com/AdamPippert/Lobstertank/internal/model"
)

// scanner abstracts sql.Row and sql.Rows for shared scanning logic.
type scanner interface {
	Scan(dest ...any) error
}

// scanGateway reads a single row into a model.Gateway.
func scanGateway(row scanner) (*model.Gateway, error) {
	var (
		gw              model.Gateway
		transportParams string
		authParams      string
		labels          string
		lastSeenAt      sql.NullTime
		ttlSeconds      sql.NullInt64
	)

	err := row.Scan(
		&gw.ID,
		&gw.Name,
		&gw.Description,
		&gw.Endpoint,
		&gw.Transport.Type,
		&transportParams,
		&gw.Auth.Type,
		&authParams,
		&gw.Auth.SecretRef,
		&gw.Status,
		&labels,
		&gw.EnrolledAt,
		&lastSeenAt,
		&ttlSeconds,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(transportParams), &gw.Transport.Params); err != nil {
		gw.Transport.Params = map[string]string{}
	}
	if err := json.Unmarshal([]byte(authParams), &gw.Auth.Params); err != nil {
		gw.Auth.Params = map[string]string{}
	}
	if err := json.Unmarshal([]byte(labels), &gw.Labels); err != nil {
		gw.Labels = map[string]string{}
	}

	if lastSeenAt.Valid {
		gw.LastSeenAt = &lastSeenAt.Time
	}
	if ttlSeconds.Valid {
		v := int(ttlSeconds.Int64)
		gw.TTLSeconds = &v
	}

	return &gw, nil
}

// gatewayColumns is the ordered column list for SELECT queries.
const gatewayColumns = `id, name, description, endpoint, transport_type, transport_params,
    auth_type, auth_params, auth_secret_ref, status, labels,
    enrolled_at, last_seen_at, ttl_seconds`

// marshalJSONMap serializes a map to a JSON string for storage.
func marshalJSONMap(m map[string]string) string {
	if m == nil {
		return "{}"
	}
	data, err := json.Marshal(m)
	if err != nil {
		return "{}"
	}
	return string(data)
}

