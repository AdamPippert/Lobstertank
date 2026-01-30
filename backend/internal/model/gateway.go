package model

import (
	"time"
)

// Status represents the health state of a gateway.
type Status string

const (
	StatusOnline   Status = "online"
	StatusOffline  Status = "offline"
	StatusDegraded Status = "degraded"
	StatusUnknown  Status = "unknown"
)

// Gateway represents a registered OpenClaw gateway instance.
type Gateway struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Endpoint    string            `json:"endpoint"`
	Transport   TransportConfig   `json:"transport"`
	Auth        GatewayAuthConfig `json:"auth"`
	Status      Status            `json:"status"`
	Labels      map[string]string `json:"labels,omitempty"`
	EnrolledAt  time.Time         `json:"enrolled_at"`
	LastSeenAt  *time.Time        `json:"last_seen_at,omitempty"`
	TTLSeconds  *int              `json:"ttl_seconds,omitempty"`
}

// TransportConfig defines how Lobstertank connects to a gateway.
type TransportConfig struct {
	Type   string            `json:"type"` // "https", "tailscale", "headscale", "cloudflare"
	Params map[string]string `json:"params,omitempty"`
}

// GatewayAuthConfig defines how Lobstertank authenticates with a gateway.
type GatewayAuthConfig struct {
	Type      string            `json:"type"` // "token", "mtls", "oidc"
	Params    map[string]string `json:"params,omitempty"`
	SecretRef string            `json:"secret_ref,omitempty"` // URI referencing a secret provider entry
}

// CreateGatewayRequest is the payload for registering a new gateway.
type CreateGatewayRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Endpoint    string            `json:"endpoint"`
	Transport   TransportConfig   `json:"transport"`
	Auth        GatewayAuthConfig `json:"auth"`
	Labels      map[string]string `json:"labels,omitempty"`
	TTLSeconds  *int              `json:"ttl_seconds,omitempty"`
}

// UpdateGatewayRequest is the payload for updating an existing gateway.
type UpdateGatewayRequest struct {
	Name        *string            `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	Endpoint    *string            `json:"endpoint,omitempty"`
	Transport   *TransportConfig   `json:"transport,omitempty"`
	Auth        *GatewayAuthConfig `json:"auth,omitempty"`
	Labels      map[string]string  `json:"labels,omitempty"`
	TTLSeconds  *int               `json:"ttl_seconds,omitempty"`
}

// HealthCheckResult is returned when probing a gateway.
type HealthCheckResult struct {
	GatewayID string `json:"gateway_id"`
	Status    Status `json:"status"`
	Latency   string `json:"latency,omitempty"`
	Error     string `json:"error,omitempty"`
	CheckedAt string `json:"checked_at"`
}
