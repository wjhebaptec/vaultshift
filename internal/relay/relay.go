// Package relay provides cross-provider secret forwarding with optional
// transformation and filtering before writing to destination providers.
package relay

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Rule describes a single forwarding rule from a source key to a destination.
type Rule struct {
	SourceProvider string
	SourceKey      string
	DestProvider   string
	DestKey        string
	Transform      func(string) (string, error) // optional value transform
}

// Result holds the outcome of applying a single rule.
type Result struct {
	Rule    Rule
	Skipped bool
	Err     error
}

// Relay forwards secrets between providers according to registered rules.
type Relay struct {
	reg   *provider.Registry
	rules []Rule
}

// New creates a Relay backed by the given provider registry.
func New(reg *provider.Registry) *Relay {
	return &Relay{reg: reg}
}

// Register adds a forwarding rule to the relay.
func (r *Relay) Register(rule Rule) error {
	if rule.SourceProvider == "" || rule.SourceKey == "" {
		return fmt.Errorf("relay: source provider and key are required")
	}
	if rule.DestProvider == "" {
		return fmt.Errorf("relay: destination provider is required")
	}
	if rule.DestKey == "" {
		rule.DestKey = rule.SourceKey
	}
	r.rules = append(r.rules, rule)
	return nil
}

// Forward applies all registered rules, returning one Result per rule.
func (r *Relay) Forward(ctx context.Context) []Result {
	results := make([]Result, 0, len(r.rules))
	for _, rule := range r.rules {
		res := r.apply(ctx, rule)
		results = append(results, res)
	}
	return results
}

func (r *Relay) apply(ctx context.Context, rule Rule) Result {
	src, ok := r.reg.Get(rule.SourceProvider)
	if !ok {
		return Result{Rule: rule, Err: fmt.Errorf("relay: unknown source provider %q", rule.SourceProvider)}
	}
	dst, ok := r.reg.Get(rule.DestProvider)
	if !ok {
		return Result{Rule: rule, Err: fmt.Errorf("relay: unknown dest provider %q", rule.DestProvider)}
	}
	val, err := src.Get(ctx, rule.SourceKey)
	if err != nil {
		return Result{Rule: rule, Err: fmt.Errorf("relay: get %q: %w", rule.SourceKey, err)}
	}
	if rule.Transform != nil {
		val, err = rule.Transform(val)
		if err != nil {
			return Result{Rule: rule, Err: fmt.Errorf("relay: transform %q: %w", rule.SourceKey, err)}
		}
	}
	if err := dst.Put(ctx, rule.DestKey, val); err != nil {
		return Result{Rule: rule, Err: fmt.Errorf("relay: put %q: %w", rule.DestKey, err)}
	}
	return Result{Rule: rule}
}

// HasFailures returns true if any result contains an error.
func HasFailures(results []Result) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}
