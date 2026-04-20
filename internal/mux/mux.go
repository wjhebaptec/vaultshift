// Package mux provides a key-based routing multiplexer that dispatches
// provider operations to different backends based on a routing function.
package mux

import (
	"context"
	"errors"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Router is a function that maps a secret key to a provider name.
type Router func(key string) string

// Mux routes provider operations to registered backends using a Router.
type Mux struct {
	router    Router
	providers map[string]provider.Provider
}

// New creates a Mux with the given routing function.
// Returns an error if router is nil.
func New(router Router) (*Mux, error) {
	if router == nil {
		return nil, errors.New("mux: router must not be nil")
	}
	return &Mux{
		router:    router,
		providers: make(map[string]provider.Provider),
	}, nil
}

// Register adds a named provider to the mux.
func (m *Mux) Register(name string, p provider.Provider) error {
	if name == "" {
		return errors.New("mux: provider name must not be empty")
	}
	if p == nil {
		return errors.New("mux: provider must not be nil")
	}
	m.providers[name] = p
	return nil
}

func (m *Mux) route(key string) (provider.Provider, error) {
	name := m.router(key)
	p, ok := m.providers[name]
	if !ok {
		return nil, fmt.Errorf("mux: no provider registered for route %q (key=%q)", name, key)
	}
	return p, nil
}

// Get retrieves a secret from the provider selected by the router.
func (m *Mux) Get(ctx context.Context, key string) (string, error) {
	p, err := m.route(key)
	if err != nil {
		return "", err
	}
	return p.Get(ctx, key)
}

// Put writes a secret to the provider selected by the router.
func (m *Mux) Put(ctx context.Context, key, value string) error {
	p, err := m.route(key)
	if err != nil {
		return err
	}
	return p.Put(ctx, key, value)
}

// Delete removes a secret from the provider selected by the router.
func (m *Mux) Delete(ctx context.Context, key string) error {
	p, err := m.route(key)
	if err != nil {
		return err
	}
	return p.Delete(ctx, key)
}

// List returns all keys from all registered providers, deduplicated.
func (m *Mux) List(ctx context.Context) ([]string, error) {
	seen := make(map[string]struct{})
	var all []string
	for _, p := range m.providers {
		keys, err := p.List(ctx)
		if err != nil {
			return nil, err
		}
		for _, k := range keys {
			if _, ok := seen[k]; !ok {
				seen[k] = struct{}{}
				all = append(all, k)
			}
		}
	}
	return all, nil
}
