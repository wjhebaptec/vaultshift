// Package propagate provides utilities for propagating secrets from a
// source provider to one or more destination providers, with optional
// key transformation applied during the copy.
package propagate

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Rule describes a single propagation mapping: a source key on a named
// provider that should be written to a destination key on another provider.
type Rule struct {
	SourceProvider string
	SourceKey      string
	DestProvider   string
	DestKey        string
}

// Propagator copies secrets between providers according to a set of Rules.
type Propagator struct {
	reg   *provider.Registry
	rules []Rule
}

// New returns a Propagator backed by the given registry.
func New(reg *provider.Registry) *Propagator {
	return &Propagator{reg: reg}
}

// AddRule appends a propagation rule to the Propagator.
func (p *Propagator) AddRule(r Rule) {
	p.rules = append(p.rules, r)
}

// Propagate executes all registered rules in order. It returns the first
// error encountered, leaving subsequent rules unattempted.
func (p *Propagator) Propagate(ctx context.Context) error {
	for _, r := range p.rules {
		if err := p.apply(ctx, r); err != nil {
			return fmt.Errorf("propagate %s/%s -> %s/%s: %w",
				r.SourceProvider, r.SourceKey,
				r.DestProvider, r.DestKey, err)
		}
	}
	return nil
}

// PropagateAll executes all registered rules and collects every error,
// returning a combined error if any rule failed.
func (p *Propagator) PropagateAll(ctx context.Context) error {
	var errs []error
	for _, r := range p.rules {
		if err := p.apply(ctx, r); err != nil {
			errs = append(errs, fmt.Errorf("%s/%s -> %s/%s: %w",
				r.SourceProvider, r.SourceKey,
				r.DestProvider, r.DestKey, err))
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("%d propagation error(s): %v", len(errs), errs)
}

func (p *Propagator) apply(ctx context.Context, r Rule) error {
	src, ok := p.reg.Get(r.SourceProvider)
	if !ok {
		return fmt.Errorf("unknown source provider %q", r.SourceProvider)
	}
	dst, ok := p.reg.Get(r.DestProvider)
	if !ok {
		return fmt.Errorf("unknown destination provider %q", r.DestProvider)
	}
	val, err := src.GetSecret(ctx, r.SourceKey)
	if err != nil {
		return fmt.Errorf("get %q: %w", r.SourceKey, err)
	}
	if err := dst.PutSecret(ctx, r.DestKey, val); err != nil {
		return fmt.Errorf("put %q: %w", r.DestKey, err)
	}
	return nil
}
