package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the complete application configuration.
type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Auth      AuthConfig
	Secrets   SecretsConfig
	Transport TransportConfig
	Audit     AuditConfig
}

// ServerConfig defines the HTTP listener settings.
type ServerConfig struct {
	Host string
	Port int
}

// DatabaseConfig defines the persistence layer settings.
type DatabaseConfig struct {
	Driver string // "postgres" or "sqlite"
	DSN    string
}

// AuthConfig defines the authentication provider settings.
type AuthConfig struct {
	Provider     string // "token" or "oidc"
	TokenSecret  string
	OIDCIssuer   string
	OIDCClientID string
	OIDCAudience string
}

// SecretsConfig defines the secret management provider settings.
type SecretsConfig struct {
	Provider      string // "builtin" or "vault"
	EncryptionKey string
	VaultAddr     string
	VaultToken    string
}

// TransportConfig defines the network transport settings.
type TransportConfig struct {
	Default string // "https", "tailscale", "headscale", "cloudflare"
}

// AuditConfig defines the audit logging settings.
type AuditConfig struct {
	Enabled bool
	Output  string // "stdout" or "file"
	Path    string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	port, err := strconv.Atoi(envOrDefault("LT_SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid LT_SERVER_PORT: %w", err)
	}

	auditEnabled, err := strconv.ParseBool(envOrDefault("LT_AUDIT_ENABLED", "true"))
	if err != nil {
		return nil, fmt.Errorf("invalid LT_AUDIT_ENABLED: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Host: envOrDefault("LT_SERVER_HOST", "0.0.0.0"),
			Port: port,
		},
		Database: DatabaseConfig{
			Driver: envOrDefault("LT_DB_DRIVER", "sqlite"),
			DSN:    envOrDefault("LT_DB_DSN", "lobstertank.db"),
		},
		Auth: AuthConfig{
			Provider:     envOrDefault("LT_AUTH_PROVIDER", "token"),
			TokenSecret:  os.Getenv("LT_AUTH_TOKEN_SECRET"),
			OIDCIssuer:   os.Getenv("LT_AUTH_OIDC_ISSUER"),
			OIDCClientID: os.Getenv("LT_AUTH_OIDC_CLIENT_ID"),
			OIDCAudience: os.Getenv("LT_AUTH_OIDC_AUDIENCE"),
		},
		Secrets: SecretsConfig{
			Provider:      envOrDefault("LT_SECRETS_PROVIDER", "builtin"),
			EncryptionKey: os.Getenv("LT_SECRETS_ENCRYPTION_KEY"),
			VaultAddr:     os.Getenv("LT_SECRETS_VAULT_ADDR"),
			VaultToken:    os.Getenv("LT_SECRETS_VAULT_TOKEN"),
		},
		Transport: TransportConfig{
			Default: envOrDefault("LT_TRANSPORT_DEFAULT", "https"),
		},
		Audit: AuditConfig{
			Enabled: auditEnabled,
			Output:  envOrDefault("LT_AUDIT_OUTPUT", "stdout"),
			Path:    os.Getenv("LT_AUDIT_PATH"),
		},
	}, nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
