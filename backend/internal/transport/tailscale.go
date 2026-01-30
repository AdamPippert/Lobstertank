package transport

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// newTailscaleClient returns an http.Client configured for Tailscale transport.
//
// When Tailscale is running on the host, connections to tailnet nodes
// (MagicDNS names like "node.tailnet.ts.net" or 100.x.y.z addresses)
// are routed transparently through the WireGuard tunnel by the Tailscale
// daemon. This transport configures appropriate timeouts and TLS settings
// for tailnet communication.
//
// Supported params:
//   - hostname: The MagicDNS hostname of the target node (informational).
//   - control_url: The Tailscale control server URL (for logging/verification).
func newTailscaleClient(params map[string]string) *http.Client {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			DialContext: dialer.DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
			MaxIdleConns:        50,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			// Tailscale connections may have higher initial latency
			// due to DERP relay fallback before direct connections establish.
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}
}
