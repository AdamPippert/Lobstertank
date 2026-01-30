package secrets

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// VaultProvider implements the secrets Provider interface using HashiCorp Vault
// KV v2 secrets engine.
type VaultProvider struct {
	addr      string
	token     string
	mountPath string
	client    *http.Client
}

// NewVaultProvider creates a Vault-backed secrets provider.
// addr is the Vault server address (e.g., "https://vault.example.com").
// token is the Vault authentication token.
// mountPath is the KV v2 mount point (e.g., "secret").
func NewVaultProvider(addr, token, mountPath string) (*VaultProvider, error) {
	if addr == "" {
		return nil, fmt.Errorf("vault address is required")
	}
	if token == "" {
		return nil, fmt.Errorf("vault token is required")
	}
	if mountPath == "" {
		mountPath = "secret"
	}

	return &VaultProvider{
		addr:      strings.TrimRight(addr, "/"),
		token:     token,
		mountPath: mountPath,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// vaultKVResponse represents the Vault KV v2 read response.
type vaultKVResponse struct {
	Data struct {
		Data map[string]interface{} `json:"data"`
	} `json:"data"`
}

// Resolve retrieves a secret from Vault KV v2.
// The ref is the path within the mount, e.g., "lobstertank/gateway-a/token".
// The secret value is stored under the "value" key in the KV data.
func (p *VaultProvider) Resolve(ctx context.Context, ref string) (string, error) {
	url := fmt.Sprintf("%s/v1/%s/data/%s", p.addr, p.mountPath, strings.TrimPrefix(ref, "/"))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build vault read request: %w", err)
	}
	req.Header.Set("X-Vault-Token", p.token)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("vault read request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read vault response: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("secret not found in vault: %s", ref)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("vault returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var kvResp vaultKVResponse
	if err := json.Unmarshal(body, &kvResp); err != nil {
		return "", fmt.Errorf("decode vault response: %w", err)
	}

	value, ok := kvResp.Data.Data["value"]
	if !ok {
		return "", fmt.Errorf("secret at %s has no 'value' key", ref)
	}

	strValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("secret value at %s is not a string", ref)
	}

	return strValue, nil
}

// Store writes a secret to Vault KV v2.
// The ref is the path within the mount. The value is stored under the "value" key.
func (p *VaultProvider) Store(ctx context.Context, ref string, value string) error {
	url := fmt.Sprintf("%s/v1/%s/data/%s", p.addr, p.mountPath, strings.TrimPrefix(ref, "/"))

	payload := map[string]interface{}{
		"data": map[string]interface{}{
			"value": value,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal vault write payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build vault write request: %w", err)
	}
	req.Header.Set("X-Vault-Token", p.token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("vault write request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("vault write returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// Delete removes a secret from Vault KV v2 by deleting all versions and metadata.
func (p *VaultProvider) Delete(ctx context.Context, ref string) error {
	url := fmt.Sprintf("%s/v1/%s/metadata/%s", p.addr, p.mountPath, strings.TrimPrefix(ref, "/"))

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("build vault delete request: %w", err)
	}
	req.Header.Set("X-Vault-Token", p.token)

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("vault delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("vault delete returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}
