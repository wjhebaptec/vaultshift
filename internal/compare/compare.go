// Package compare provides utilities for comparing secret values across
// multiple providers, returning a structured result of matches and mismatches.
package compare

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Result holds the outcome of comparing a single secret key across providers.
type Result struct {
	Key      string
	Match    bool
	Values   map[string]string // provider name -> value
	Missing  []string          // provider names where key was absent
}

// Comparer compares secret values across a set of named providers.
type Comparer struct {
	registry *provider.Registry
}

// New returns a new Comparer backed by the given registry.
func New(r *provider.Registry) (*Comparer, error) {
	if r == nil {
		return nil, fmt.Errorf("compare: registry must not be nil")
	}
	return &Comparer{registry: r}, nil
}

// Compare fetches the given key from each of the named providers and returns
// a Result describing whether all present values agree.
func (c *Comparer) Compare(ctx context.Context, key string, providerNames []string) (Result, error) {
	if key == "" {
		return Result{}, fmt.Errorf("compare: key must not be empty")
	}
	if len(providerNames) == 0 {
		return Result{}, fmt.Errorf("compare: at least one provider name required")
	}

	res := Result{
		Key:    key,
		Values: make(map[string]string),
	}

	for _, name := range providerNames {
		p, err := c.registry.Get(name)
		if err != nil {
			return Result{}, fmt.Errorf("compare: provider %q not found: %w", name, err)
		}
		val, err := p.GetSecret(ctx, key)
		if err != nil {
			res.Missing = append(res.Missing, name)
			continue
		}
		res.Values[name] = val
	}

	res.Match = len(res.Missing) == 0 && allEqual(res.Values)
	return res, nil
}

// CompareAll compares every key in keys across the given providers.
func (c *Comparer) CompareAll(ctx context.Context, keys []string, providerNames []string) ([]Result, error) {
	results := make([]Result, 0, len(keys))
	for _, k := range keys {
		r, err := c.Compare(ctx, k, providerNames)
		if err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, nil
}

// allEqual returns true when all values in the map are identical.
func allEqual(m map[string]string) bool {
	var ref string
	first := true
	for _, v := range m {
		if first {
			ref = v
			first = false
			continue
		}
		if v != ref {
			return false
		}
	}
	return true
}
