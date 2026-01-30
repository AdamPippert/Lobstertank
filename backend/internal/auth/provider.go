package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/AdamPippert/Lobstertank/internal/config"
	"github.com/AdamPippert/Lobstertank/internal/secrets"
)

// Principal represents an authenticated identity.
type Principal struct {
	Subject string
	Roles   []string
}

type contextKey string

const principalKey contextKey = "auth.principal"

// PrincipalFromContext extracts the authenticated principal from a request context.
func PrincipalFromContext(ctx context.Context) (*Principal, bool) {
	p, ok := ctx.Value(principalKey).(*Principal)
	return p, ok
}

// ContextWithPrincipal stores a principal in the context.
func ContextWithPrincipal(ctx context.Context, p *Principal) context.Context {
	return context.WithValue(ctx, principalKey, p)
}

// Provider verifies inbound requests and extracts the caller identity.
type Provider interface {
	// Authenticate inspects the request and returns the authenticated principal.
	// Returns an error if authentication fails.
	Authenticate(ctx context.Context, r *http.Request) (*Principal, error)
}

// NewProvider constructs the appropriate auth provider based on configuration.
func NewProvider(cfg config.AuthConfig, sp secrets.Provider) (Provider, error) {
	switch cfg.Provider {
	case "token":
		if cfg.TokenSecret == "" {
			return nil, fmt.Errorf("LT_AUTH_TOKEN_SECRET is required when auth provider is 'token'")
		}
		return NewTokenProvider(cfg.TokenSecret), nil
	case "oidc":
		if cfg.OIDCIssuer == "" {
			return nil, fmt.Errorf("LT_AUTH_OIDC_ISSUER is required when auth provider is 'oidc'")
		}
		if cfg.OIDCClientID == "" {
			return nil, fmt.Errorf("LT_AUTH_OIDC_CLIENT_ID is required when auth provider is 'oidc'")
		}
		return NewOIDCProvider(context.Background(), cfg.OIDCIssuer, cfg.OIDCClientID, cfg.OIDCAudience)
	default:
		return nil, fmt.Errorf("unknown auth provider: %s", cfg.Provider)
	}
}
