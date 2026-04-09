package mock

import (
	"context"
	"fmt"
	"sync"

	"vaultshift/internal/provider"
)

// Provider is a mock implementation for testing
type Provider struct {
	mu       sync.RWMutex
	secrets  map[string]*provider.Secret
	provType provider.Type
	closed   bool
}

// New creates a new mock provider
func New(provType provider.Type) *Provider {
	return &Provider{
		secrets:  make(map[string]*provider.Secret),
		provType: provType,
	}
}

// GetSecret retrieves a secret by key
func (p *Provider) GetSecret(ctx context.Context, key string) (*provider.Secret, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, fmt.Errorf("provider is closed")
	}

	s, ok := p.secrets[key]
	if !ok {
		return nil, fmt.Errorf("secret %s not found", key)
	}
	return s, nil
}

// PutSecret stores or updates a secret
func (p *Provider) PutSecret(ctx context.Context, secret *provider.Secret) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("provider is closed")
	}

	p.secrets[secret.Key] = secret
	return nil
}

// DeleteSecret removes a secret by key
func (p *Provider) DeleteSecret(ctx context.Context, key string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return fmt.Errorf("provider is closed")
	}

	delete(p.secrets, key)
	return nil
}

// ListSecrets returns all secret keys
func (p *Provider) ListSecrets(ctx context.Context) ([]string, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.closed {
		return nil, fmt.Errorf("provider is closed")
	}

	keys := make([]string, 0, len(p.secrets))
	for k := range p.secrets {
		keys = append(keys, k)
	}
	return keys, nil
}

// Type returns the provider type
func (p *Provider) Type() provider.Type {
	return p.provType
}

// Close marks the provider as closed
func (p *Provider) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.closed = true
	return nil
}
