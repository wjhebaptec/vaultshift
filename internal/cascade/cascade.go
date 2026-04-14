// Package cascade provides a write-through secret propagation mechanism
// that writes a secret to a primary provider and fans out to one or more
// secondary providers in order, stopping on the first failure.
package cascade

import (
	"context"
	"errors"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// ErrNoProviders is returned when no providers are registered.
var ErrNoProviders = errors.New("cascade: at least one provider is required")

// Result holds the outcome of a single provider write.
type Result struct {
	Provider string
	Err      error
}

// Cascade writes a secret to a primary provider then propagates to
// subsequent providers in registration order.
type Cascade struct {
	primary    string
	chain      []string
	registry   *provider.Registry
	stopOnFail bool
}

// Option configures a Cascade.
type Option func(*Cascade)

// WithStopOnFailure causes the cascade to halt propagation when any
// secondary provider write fails.
func WithStopOnFailure() Option {
	return func(c *Cascade) { c.stopOnFail = true }
}

// New creates a Cascade with the given primary provider name, an ordered
// list of secondary provider names, and a populated registry.
func New(primary string, chain []string, reg *provider.Registry, opts ...Option) (*Cascade, error) {
	if primary == "" {
		return nil, errors.New("cascade: primary provider name must not be empty")
	}
	if len(chain) == 0 {
		return nil, ErrNoProviders
	}
	c := &Cascade{primary: primary, chain: chain, registry: reg}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

// Put writes value to the primary provider and then to each secondary
// provider in order. Results for every attempted write are returned.
func (c *Cascade) Put(ctx context.Context, key, value string) ([]Result, error) {
	prim, err := c.registry.Get(c.primary)
	if err != nil {
		return nil, fmt.Errorf("cascade: primary provider %q not found: %w", c.primary, err)
	}
	if err := prim.Put(ctx, key, value); err != nil {
		return []Result{{Provider: c.primary, Err: err}}, err
	}
	results := []Result{{Provider: c.primary}}
	for _, name := range c.chain {
		p, err := c.registry.Get(name)
		if err != nil {
			r := Result{Provider: name, Err: fmt.Errorf("cascade: provider %q not found: %w", name, err)}
			results = append(results, r)
			if c.stopOnFail {
				break
			}
			continue
		}
		wErr := p.Put(ctx, key, value)
		results = append(results, Result{Provider: name, Err: wErr})
		if wErr != nil && c.stopOnFail {
			break
		}
	}
	return results, nil
}

// HasFailures reports whether any result in the slice contains an error.
func HasFailures(results []Result) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}
