package transport

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// newHeadscaleClient returns an http.Client configured for Headscale transport.
//
// Headscale is a self-hosted implementation of the Tailscale control server.
// When the host is registered with a Headscale instance, connections to
// other nodes in the network are routed through the WireGuard mesh, similar
// to Tailscale. This transport configures the HTTP client for communication
// with gateways reachable through the Headscale network.
//
// Supported params:
//   - api_url:   The Headscale server API URL (e.g., "https://headscale.example.com").
//   - api_key:   API key for authenticating with the Headscale control server.
//   - node_name: The target node's registered name in Headscale.
func newHeadscaleClient(params map[string]string) *http.Client {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// When a custom API URL is provided, we may need to trust self-signed
	// certificates for the Headscale control server. In production, this
	// should be handled via system trust store configuration.
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			TLSClientConfig:      tlsConfig,
			MaxIdleConns:          50,
			MaxIdleConnsPerHost:   10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}
}
