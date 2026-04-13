// Package inject provides a mechanism for injecting resolved secrets
// into structured targets such as maps and environment-style configs.
package inject

import (
	"context"
	"fmt"
	"strings"
)

// Provider is the minimal interface required to fetch a secret value.
type Provider interface {
	Get(ctx context.Context, key string) (string, error)
}

// Target represents a destination that receives injected key/value pairs.
type Target interface {
	Set(key, value string)
}

// MapTarget wraps a map[string]string as a Target.
type MapTarget map[string]string

func (m MapTarget) Set(key, value string) { m[key] = value }

// Injector resolves secret keys from a Provider and writes them to a Target.
type Injector struct {
	provider Provider
	prefix   string
}

// Option configures an Injector.
type Option func(*Injector)

// WithPrefix strips the given prefix from resolved keys before writing to the target.
func WithPrefix(prefix string) Option {
	return func(i *Injector) { i.prefix = prefix }
}

// New creates a new Injector backed by the given Provider.
func New(p Provider, opts ...Option) (*Injector, error) {
	if p == nil {
		return nil, fmt.Errorf("inject: provider must not be nil")
	}
	inj := &Injector{provider: p}
	for _, o := range opts {
		o(inj)
	}
	return inj, nil
}

// Inject resolves each key from the provider and writes the result to target.
// Keys that cannot be resolved are returned as errors; injection continues for
// the remaining keys.
func (inj *Injector) Inject(ctx context.Context, keys []string, target Target) []error {
	var errs []error
	for _, k := range keys {
		val, err := inj.provider.Get(ctx, k)
		if err != nil {
			errs = append(errs, fmt.Errorf("inject: key %q: %w", k, err))
			continue
		}
		destKey := k
		if inj.prefix != "" {
			destKey = strings.TrimPrefix(k, inj.prefix)
		}
		target.Set(destKey, val)
	}
	return errs
}

// InjectMap is a convenience wrapper that injects into a fresh map and returns it.
func (inj *Injector) InjectMap(ctx context.Context, keys []string) (map[string]string, []error) {
	m := make(MapTarget, len(keys))
	errs := inj.Inject(ctx, keys, m)
	return map[string]string(m), errs
}
