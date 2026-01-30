package gateway

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AdamPippert/Lobstertank/internal/model"
	"github.com/AdamPippert/Lobstertank/internal/secrets"
	"github.com/AdamPippert/Lobstertank/internal/transport"
)

// Client communicates with a single OpenClaw gateway instance.
type Client struct {
	gateway    *model.Gateway
	httpClient *http.Client
	secretProv secrets.Provider
}

// ClientFactory creates gateway clients configured with the correct transport
// and authentication.
type ClientFactory struct {
	transport  transport.Provider
	secretProv secrets.Provider
}

// NewClientFactory returns a factory that builds gateway clients.
func NewClientFactory(tp transport.Provider, sp secrets.Provider) *ClientFactory {
	return &ClientFactory{transport: tp, secretProv: sp}
}

// ClientFor builds a Client configured for the given gateway.
func (f *ClientFactory) ClientFor(gw *model.Gateway) *Client {
	httpClient := f.transport.HTTPClient(gw.Transport.Type, gw.Transport.Params)
	return &Client{
		gateway:    gw,
		httpClient: httpClient,
		secretProv: f.secretProv,
	}
}

// HealthCheck probes the gateway and returns its status.
func (c *Client) HealthCheck(ctx context.Context) (*model.HealthCheckResult, error) {
	start := time.Now()
	result := &model.HealthCheckResult{
		GatewayID: c.gateway.ID,
		CheckedAt: start.UTC().Format(time.RFC3339),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.gateway.Endpoint+"/healthz", nil)
	if err != nil {
		result.Status = model.StatusOffline
		result.Error = err.Error()
		return result, fmt.Errorf("build health request: %w", err)
	}

	if err := c.applyAuth(ctx, req); err != nil {
		result.Status = model.StatusUnknown
		result.Error = "auth setup failed"
		return result, fmt.Errorf("apply auth: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		result.Status = model.StatusOffline
		result.Error = err.Error()
		result.Latency = time.Since(start).String()
		return result, nil // Not an application error — gateway is simply unreachable.
	}
	defer resp.Body.Close()

	result.Latency = time.Since(start).String()

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		result.Status = model.StatusOnline
	case resp.StatusCode >= 500:
		result.Status = model.StatusDegraded
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	default:
		result.Status = model.StatusUnknown
		result.Error = fmt.Sprintf("HTTP %d", resp.StatusCode)
	}

	return result, nil
}

// SendPrompt sends a prompt to the OpenClaw gateway and returns the raw response body.
// This is a stub — the actual OpenClaw API contract will define the real implementation.
func (c *Client) SendPrompt(ctx context.Context, prompt string) ([]byte, error) {
	// TODO: Implement once OpenClaw API contract is defined.
	_ = ctx
	_ = prompt
	return nil, fmt.Errorf("SendPrompt not yet implemented for gateway %s", c.gateway.ID)
}

// applyAuth adds authentication headers or TLS config to the outbound request.
func (c *Client) applyAuth(ctx context.Context, req *http.Request) error {
	switch c.gateway.Auth.Type {
	case "token":
		token, err := c.resolveSecret(ctx, c.gateway.Auth.SecretRef)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+token)
	case "mtls":
		// mTLS is handled at the transport/TLS layer; no header needed.
	case "oidc":
		// TODO: Implement OIDC token acquisition.
	default:
		// No auth configured.
	}
	return nil
}

func (c *Client) resolveSecret(ctx context.Context, ref string) (string, error) {
	if ref == "" {
		// Fall back to inline param if no secret ref is set.
		if tok, ok := c.gateway.Auth.Params["token"]; ok {
			return tok, nil
		}
		return "", fmt.Errorf("no secret reference or inline token for gateway %s", c.gateway.ID)
	}
	return c.secretProv.Resolve(ctx, ref)
}
