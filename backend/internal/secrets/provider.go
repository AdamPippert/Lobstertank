package secrets

import (
	"context"
	"fmt"

	"github.com/AdamPippert/Lobstertank/internal/config"
)

// Provider abstracts secret storage and retrieval. Secrets are referenced by
// URI-style keys (e.g., "builtin://gateway-a/token" or "vault://secret/data/gw-a").
type Provider interface {
	// Resolve retrieves the plaintext value for the given secret reference.
	Resolve(ctx context.Context, ref string) (string, error)

	// Store persists a secret under the given reference.
	Store(ctx context.Context, ref string, value string) error

	// Delete removes a secret by reference.
	Delete(ctx context.Context, ref string) error
}

// NewProvider constructs the appropriate secrets provider based on configuration.
func NewProvider(cfg config.SecretsConfig) (Provider, error) {
	switch cfg.Provider {
	case "builtin":
		return NewBuiltinProvider(cfg.EncryptionKey)
	case "vault":
		// TODO: Implement HashiCorp Vault provider.
		return nil, fmt.Errorf("vault secrets provider is not yet implemented")
	default:
		return nil, fmt.Errorf("unknown secrets provider: %s", cfg.Provider)
	}
}
