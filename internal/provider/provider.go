package provider

import (
	"context"
	"fmt"
)

// Type represents the type of secret provider
type Type string

const (
	TypeAWS   Type = "aws"
	TypeGCP   Type = "gcp"
	TypeVault Type = "vault"
)

// Secret represents a secret value with metadata
type Secret struct {
	Key      string
	Value    string
	Version  string
	Metadata map[string]string
}

// Provider defines the interface for secret management operations
type Provider interface {
	// GetSecret retrieves a secret by key
	GetSecret(ctx context.Context, key string) (*Secret, error)
	
	// PutSecret stores or updates a secret
	PutSecret(ctx context.Context, secret *Secret) error
	
	// DeleteSecret removes a secret by key
	DeleteSecret(ctx context.Context, key string) error
	
	// ListSecrets returns all secret keys
	ListSecrets(ctx context.Context) ([]string, error)
	
	// Type returns the provider type
	Type() Type
	
	// Close cleans up provider resources
	Close() error
}

// Config holds common provider configuration
type Config struct {
	Type   Type
	Region string
	Prefix string
}

// Registry manages provider instances
type Registry struct {
	providers map[Type]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[Type]Provider),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(p Provider) {
	r.providers[p.Type()] = p
}

// Get retrieves a provider by type
func (r *Registry) Get(t Type) (Provider, error) {
	p, ok := r.providers[t]
	if !ok {
		return nil, fmt.Errorf("provider %s not registered", t)
	}
	return p, nil
}

// Close closes all registered providers
func (r *Registry) Close() error {
	for _, p := range r.providers {
		if err := p.Close(); err != nil {
			return err
		}
	}
	return nil
}
