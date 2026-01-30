package secrets

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
)

// BuiltinProvider stores secrets in memory with AES-GCM encryption.
// In production, the encrypted map should be persisted to the data store.
type BuiltinProvider struct {
	mu      sync.RWMutex
	secrets map[string]string // ref -> base64(encrypted value)
	aead    cipher.AEAD
}

// NewBuiltinProvider creates an in-memory secrets provider using the given
// base64-encoded 32-byte AES key.
func NewBuiltinProvider(encKeyBase64 string) (*BuiltinProvider, error) {
	if encKeyBase64 == "" {
		// Allow startup without encryption for development.
		return &BuiltinProvider{
			secrets: make(map[string]string),
		}, nil
	}

	keyBytes, err := base64.StdEncoding.DecodeString(encKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("decode encryption key: %w", err)
	}
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("encryption key must be 32 bytes, got %d", len(keyBytes))
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("create AES cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("create GCM: %w", err)
	}

	return &BuiltinProvider{
		secrets: make(map[string]string),
		aead:    aead,
	}, nil
}

// Resolve decrypts and returns the secret for the given reference.
func (p *BuiltinProvider) Resolve(_ context.Context, ref string) (string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	enc, ok := p.secrets[ref]
	if !ok {
		return "", fmt.Errorf("secret not found: %s", ref)
	}

	if p.aead == nil {
		// No encryption configured — return raw value.
		return enc, nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(enc)
	if err != nil {
		return "", fmt.Errorf("decode secret: %w", err)
	}

	nonceSize := p.aead.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := p.aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt secret: %w", err)
	}

	return string(plaintext), nil
}

// Store encrypts and saves a secret under the given reference.
func (p *BuiltinProvider) Store(_ context.Context, ref string, value string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.aead == nil {
		p.secrets[ref] = value
		return nil
	}

	nonce := make([]byte, p.aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("generate nonce: %w", err)
	}

	ciphertext := p.aead.Seal(nonce, nonce, []byte(value), nil)
	p.secrets[ref] = base64.StdEncoding.EncodeToString(ciphertext)
	return nil
}

// Delete removes a secret by reference.
func (p *BuiltinProvider) Delete(_ context.Context, ref string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.secrets, ref)
	return nil
}
