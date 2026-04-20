// Package partition splits secrets across multiple providers based on a
// routing function, allowing different keys to be stored in different backends.
package partition

import (
	"context"
	"errors"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// RouterFunc decides which provider name a given key should be routed to.
type RouterFunc func(key string) string

// Partitioner routes secret operations to providers based on a key routing function.
type Partitioner struct {
	router    RouterFunc
	providers map[string]provider.Provider
}

// New creates a Partitioner with the given router and provider registry.
// Returns an error if router is nil or no providers are supplied.
func New(router RouterFunc, providers map[string]provider.Provider) (*Partitioner, error) {
	if router == nil {
		return nil, errors.New("partition: router must not be nil")
	}
	if len(providers) == 0 {
		return nil, errors.New("partition: at least one provider is required")
	}
	return &Partitioner{router: router, providers: providers}, nil
}

func (p *Partitioner) resolve(key string) (provider.Provider, error) {
	name := p.router(key)
	prov, ok := p.providers[name]
	if !ok {
		return nil, fmt.Errorf("partition: no provider registered for name %q (key=%q)", name, key)
	}
	return prov, nil
}

// Put stores the secret value in the provider selected by the router.
func (p *Partitioner) Put(ctx context.Context, key, value string) error {
	prov, err := p.resolve(key)
	if err != nil {
		return err
	}
	return prov.Put(ctx, key, value)
}

// Get retrieves the secret value from the provider selected by the router.
func (p *Partitioner) Get(ctx context.Context, key string) (string, error) {
	prov, err := p.resolve(key)
	if err != nil {
		return "", err
	}
	return prov.Get(ctx, key)
}

// Delete removes the secret from the provider selected by the router.
func (p *Partitioner) Delete(ctx context.Context, key string) error {
	prov, err := p.resolve(key)
	if err != nil {
		return err
	}
	return prov.Delete(ctx, key)
}

// Providers returns the names of all registered providers.
func (p *Partitioner) Providers() []string {
	names := make([]string, 0, len(p.providers))
	for name := range p.providers {
		names = append(names, name)
	}
	return names
}
