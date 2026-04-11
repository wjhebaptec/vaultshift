// Package resolve provides secret key resolution with aliasing and
// fallback chain support across multiple providers.
package resolve

import (
	"context"
	"errors"
	"fmt"
)

// ErrNotResolved is returned when no provider in the chain can resolve a key.
var ErrNotResolved = errors.New("resolve: key not found in any provider")

// Provider is the minimal interface required for resolution.
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
}

// Alias maps a logical name to a concrete secret key.
type Alias struct {
	Name string
	Key  string
}

// Resolver resolves secret keys using aliases and a fallback provider chain.
type Resolver struct {
	providers []Provider
	aliases   map[string]string
}

// Option configures a Resolver.
type Option func(*Resolver)

// WithAlias registers a logical name that maps to a concrete key.
func WithAlias(name, key string) Option {
	return func(r *Resolver) {
		r.aliases[name] = key
	}
}

// New creates a Resolver that queries providers in order until a value is found.
func New(providers []Provider, opts ...Option) *Resolver {
	r := &Resolver{
		providers: providers,
		aliases:   make(map[string]string),
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Resolve returns the value for key, consulting aliases first then walking
// the provider chain. Returns ErrNotResolved if no provider has the key.
func (r *Resolver) Resolve(ctx context.Context, key string) (string, error) {
	resolved := key
	if mapped, ok := r.aliases[key]; ok {
		resolved = mapped
	}
	for _, p := range r.providers {
		val, err := p.Get(ctx, resolved)
		if err == nil {
			return val, nil
		}
	}
	return "", fmt.Errorf("%w: %s", ErrNotResolved, key)
}

// ResolveAll resolves a slice of keys, returning a map of key→value.
// Keys that cannot be resolved are omitted and their errors collected.
func (r *Resolver) ResolveAll(ctx context.Context, keys []string) (map[string]string, []error) {
	out := make(map[string]string, len(keys))
	var errs []error
	for _, k := range keys {
		v, err := r.Resolve(ctx, k)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		out[k] = v
	}
	return out, errs
}
