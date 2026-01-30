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
