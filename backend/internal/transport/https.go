package transport

import (
	"crypto/tls"
	"net/http"
	"time"
)

// newHTTPSClient returns a standard HTTPS client with sensible timeouts and
// TLS defaults.
func newHTTPSClient(_ map[string]string) *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		},
	}
}

// newTailscaleClient returns an http.Client configured for Tailscale transport.
// TODO: Integrate with tsnet or Tailscale local API for direct node dialing.
func newTailscaleClient(_ map[string]string) *http.Client {
	// Stub — falls back to standard HTTPS for now.
	return newHTTPSClient(nil)
}

// newHeadscaleClient returns an http.Client configured for Headscale transport.
// TODO: Integrate with Headscale API for node resolution and dialing.
func newHeadscaleClient(_ map[string]string) *http.Client {
	// Stub — falls back to standard HTTPS for now.
	return newHTTPSClient(nil)
}

// newCloudflareClient returns an http.Client configured for Cloudflare Tunnel transport.
// TODO: Integrate with cloudflared tunnel client for routed connections.
func newCloudflareClient(_ map[string]string) *http.Client {
	// Stub — falls back to standard HTTPS for now.
	return newHTTPSClient(nil)
}
