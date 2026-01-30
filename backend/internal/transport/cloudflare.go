package transport

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"
)

// cfAccessTransport wraps an http.RoundTripper to inject Cloudflare Access
// service token headers on every outbound request.
type cfAccessTransport struct {
	base          http.RoundTripper
	clientID      string
	clientSecret  string
}

func (t *cfAccessTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid mutating the original.
	r := req.Clone(req.Context())
	if t.clientID != "" {
		r.Header.Set("CF-Access-Client-Id", t.clientID)
	}
	if t.clientSecret != "" {
		r.Header.Set("CF-Access-Client-Secret", t.clientSecret)
	}
	return t.base.RoundTrip(r)
}

// newCloudflareClient returns an http.Client configured for Cloudflare Tunnel
// and Cloudflare Access transport.
//
// When a gateway is exposed behind Cloudflare Tunnel (via cloudflared), the
// client must present valid Cloudflare Access service token credentials in
// request headers. This transport automatically injects the CF-Access-Client-Id
// and CF-Access-Client-Secret headers.
//
// Supported params:
//   - service_token_id:     Cloudflare Access service token client ID.
//   - service_token_secret: Cloudflare Access service token client secret.
//   - tunnel_url:           The public Cloudflare Tunnel URL (informational).
func newCloudflareClient(params map[string]string) *http.Client {
	base := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	}

	transport := http.RoundTripper(base)

	clientID := params["service_token_id"]
	clientSecret := params["service_token_secret"]
	if clientID != "" || clientSecret != "" {
		transport = &cfAccessTransport{
			base:         base,
			clientID:     clientID,
			clientSecret: clientSecret,
		}
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}
}
