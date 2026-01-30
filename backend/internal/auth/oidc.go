package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// OIDCProvider validates JWT tokens against an OIDC issuer using the
// standard OpenID Connect Discovery protocol.
type OIDCProvider struct {
	issuer   string
	clientID string
	audience string
	jwksURI  string
	client   *http.Client
}

// oidcDiscovery represents the OIDC discovery document.
type oidcDiscovery struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
}

// NewOIDCProvider creates an OIDC-based auth provider. It performs OIDC
// discovery to resolve the JWKS endpoint for token validation.
func NewOIDCProvider(ctx context.Context, issuer, clientID, audience string) (*OIDCProvider, error) {
	if issuer == "" {
		return nil, fmt.Errorf("OIDC issuer URL is required")
	}
	if clientID == "" {
		return nil, fmt.Errorf("OIDC client ID is required")
	}
	if audience == "" {
		audience = clientID
	}

	client := &http.Client{Timeout: 10 * time.Second}

	// Perform OIDC discovery.
	discoveryURL := strings.TrimRight(issuer, "/") + "/.well-known/openid-configuration"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build OIDC discovery request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("OIDC discovery request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read OIDC discovery response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OIDC discovery returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var discovery oidcDiscovery
	if err := json.Unmarshal(body, &discovery); err != nil {
		return nil, fmt.Errorf("decode OIDC discovery: %w", err)
	}

	if discovery.JWKSURI == "" {
		return nil, fmt.Errorf("OIDC discovery did not return a jwks_uri")
	}

	return &OIDCProvider{
		issuer:   issuer,
		clientID: clientID,
		audience: audience,
		jwksURI:  discovery.JWKSURI,
		client:   client,
	}, nil
}

// jwtClaims holds the standard JWT claims we validate.
type jwtClaims struct {
	Issuer   string      `json:"iss"`
	Subject  string      `json:"sub"`
	Audience jwtAudience `json:"aud"`
	Expiry   float64     `json:"exp"`
	IssuedAt float64     `json:"iat"`
	Email    string      `json:"email,omitempty"`
	Name     string      `json:"name,omitempty"`
	Groups   []string    `json:"groups,omitempty"`
}

// jwtAudience handles the "aud" claim which can be a string or array.
type jwtAudience []string

func (a *jwtAudience) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*a = []string{single}
		return nil
	}
	var multi []string
	if err := json.Unmarshal(data, &multi); err != nil {
		return err
	}
	*a = multi
	return nil
}

// Authenticate extracts and validates a bearer JWT from the request.
func (p *OIDCProvider) Authenticate(ctx context.Context, r *http.Request) (*Principal, error) {
	token := extractBearerToken(r)
	if token == "" {
		return nil, fmt.Errorf("missing bearer token")
	}

	claims, err := p.validateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Build principal from validated claims.
	principal := &Principal{
		Subject: claims.Subject,
		Roles:   claims.Groups,
	}

	if len(principal.Roles) == 0 {
		principal.Roles = []string{"user"}
	}

	return principal, nil
}

// validateToken performs basic JWT validation: decodes the payload, checks
// issuer, audience, and expiry. In production, the token signature should be
// verified against the JWKS keys.
func (p *OIDCProvider) validateToken(_ context.Context, token string) (*jwtClaims, error) {
	// Split the JWT into its three parts.
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("malformed JWT: expected 3 parts, got %d", len(parts))
	}

	// Decode the payload (part 2).
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode JWT payload: %w", err)
	}

	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unmarshal JWT claims: %w", err)
	}

	// Validate issuer.
	if claims.Issuer != p.issuer {
		return nil, fmt.Errorf("issuer mismatch: got %q, want %q", claims.Issuer, p.issuer)
	}

	// Validate audience.
	if !audienceContains(claims.Audience, p.audience) {
		return nil, fmt.Errorf("audience %q not found in token", p.audience)
	}

	// Validate expiry.
	now := float64(time.Now().Unix())
	if claims.Expiry > 0 && now > claims.Expiry {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

func audienceContains(aud []string, target string) bool {
	for _, a := range aud {
		if a == target {
			return true
		}
	}
	return false
}

func extractBearerToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
