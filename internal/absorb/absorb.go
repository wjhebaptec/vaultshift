// Package absorb provides a provider decorator that merges secrets from
// multiple source providers into a single destination provider on demand.
package absorb

import (
	"context"
	"errors"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Result holds the outcome of a single absorb operation.
type Result struct {
	Provider string
	Key      string
	Err      error
}

// Absorber copies secrets from one or more source providers into a destination.
type Absorber struct {
	reg  *provider.Registry
	dest string
}

// New creates an Absorber that writes absorbed secrets into dest.
func New(reg *provider.Registry, dest string) (*Absorber, error) {
	if reg == nil {
		return nil, errors.New("absorb: registry must not be nil")
	}
	if dest == "" {
		return nil, errors.New("absorb: dest provider name must not be empty")
	}
	return &Absorber{reg: reg, dest: dest}, nil
}

// Absorb copies the given key from src into the destination provider.
func (a *Absorber) Absorb(ctx context.Context, src, key string) error {
	srcP, err := a.reg.Get(src)
	if err != nil {
		return fmt.Errorf("absorb: source provider %q: %w", src, err)
	}
	dstP, err := a.reg.Get(a.dest)
	if err != nil {
		return fmt.Errorf("absorb: dest provider %q: %w", a.dest, err)
	}
	val, err := srcP.Get(ctx, key)
	if err != nil {
		return fmt.Errorf("absorb: get %q from %q: %w", key, src, err)
	}
	if err := dstP.Put(ctx, key, val); err != nil {
		return fmt.Errorf("absorb: put %q into %q: %w", key, a.dest, err)
	}
	return nil
}

// AbsorbAll copies all keys listed under src into the destination provider.
// Results are collected and returned; processing continues on error.
func (a *Absorber) AbsorbAll(ctx context.Context, src string) []Result {
	srcP, err := a.reg.Get(src)
	if err != nil {
		return []Result{{Provider: src, Err: fmt.Errorf("absorb: source provider %q: %w", src, err)}}
	}
	keys, err := srcP.List(ctx)
	if err != nil {
		return []Result{{Provider: src, Err: fmt.Errorf("absorb: list %q: %w", src, err)}}
	}
	results := make([]Result, 0, len(keys))
	for _, k := range keys {
		e := a.Absorb(ctx, src, k)
		results = append(results, Result{Provider: src, Key: k, Err: e})
	}
	return results
}

// HasFailures returns true if any result contains a non-nil error.
func HasFailures(results []Result) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}
