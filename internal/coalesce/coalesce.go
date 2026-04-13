// Package coalesce provides a provider that returns the first non-empty
// value found across an ordered list of secret providers.
package coalesce

import (
	"context"
	"errors"
	"fmt"
)

// Provider is the minimal interface required by the coalescer.
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
}

// ErrNotFound is returned when no provider contains a non-empty value.
var ErrNotFound = errors.New("coalesce: key not found in any provider")

// Coalescer queries providers in order and returns the first non-empty value.
type Coalescer struct {
	providers []namedProvider
}

type namedProvider struct {
	name string
	p    Provider
}

// New creates a Coalescer. Providers are queried in the order they are added.
func New() *Coalescer {
	return &Coalescer{}
}

// Add registers a provider under a logical name.
func (c *Coalescer) Add(name string, p Provider) error {
	if name == "" {
		return errors.New("coalesce: provider name must not be empty")
	}
	if p == nil {
		return errors.New("coalesce: provider must not be nil")
	}
	c.providers = append(c.providers, namedProvider{name: name, p: p})
	return nil
}

// Get returns the first non-empty value for key across all registered providers.
// It records which provider resolved the key in the returned source string.
func (c *Coalescer) Get(ctx context.Context, key string) (value, source string, err error) {
	if len(c.providers) == 0 {
		return "", "", ErrNotFound
	}
	for _, np := range c.providers {
		v, e := np.p.Get(ctx, key)
		if e != nil {
			continue
		}
		if v == "" {
			continue
		}
		return v, np.name, nil
	}
	return "", "", fmt.Errorf("%w: %s", ErrNotFound, key)
}

// GetAll resolves key from every provider and returns a map of provider name
// to value, including only providers that returned a non-empty result.
func (c *Coalescer) GetAll(ctx context.Context, key string) map[string]string {
	out := make(map[string]string)
	for _, np := range c.providers {
		v, e := np.p.Get(ctx, key)
		if e == nil && v != "" {
			out[np.name] = v
		}
	}
	return out
}
