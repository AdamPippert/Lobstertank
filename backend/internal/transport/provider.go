package transport

import (
	"net/http"

	"github.com/AdamPippert/Lobstertank/internal/config"
)

// Provider abstracts how Lobstertank establishes network connections to
// gateways. Implementations handle transport-specific concerns (TLS config,
// Tailscale dial, Cloudflare tunnel routing) and return a configured
// http.Client ready for use.
type Provider interface {
	// HTTPClient returns an http.Client configured for the given transport type
	// and parameters. If the transport type is unrecognized, a default HTTPS
	// client is returned.
	HTTPClient(transportType string, params map[string]string) *http.Client
}

// NewProvider returns the appropriate transport provider based on config.
func NewProvider(cfg config.TransportConfig) Provider {
	return &multiProvider{defaultType: cfg.Default}
}

// multiProvider delegates to the correct transport based on type.
type multiProvider struct {
	defaultType string
}

func (m *multiProvider) HTTPClient(transportType string, params map[string]string) *http.Client {
	if transportType == "" {
		transportType = m.defaultType
	}

	switch transportType {
	case "tailscale":
		return newTailscaleClient(params)
	case "headscale":
		return newHeadscaleClient(params)
	case "cloudflare":
		return newCloudflareClient(params)
	default:
		return newHTTPSClient(params)
	}
}
