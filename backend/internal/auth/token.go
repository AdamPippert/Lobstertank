package auth

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net/http"
	"strings"
)

// TokenProvider implements bearer-token authentication.
type TokenProvider struct {
	secret string
}

// NewTokenProvider creates a TokenProvider that validates against the given secret.
func NewTokenProvider(secret string) *TokenProvider {
	return &TokenProvider{secret: secret}
}

// Authenticate extracts and validates a Bearer token from the Authorization header.
func (p *TokenProvider) Authenticate(_ context.Context, r *http.Request) (*Principal, error) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, fmt.Errorf("missing Authorization header")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return nil, fmt.Errorf("invalid Authorization header format")
	}

	token := parts[1]
	if subtle.ConstantTimeCompare([]byte(token), []byte(p.secret)) != 1 {
		return nil, fmt.Errorf("invalid token")
	}

	return &Principal{
		Subject: "token-user",
		Roles:   []string{"admin"},
	}, nil
}
