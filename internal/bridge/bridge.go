// Package bridge provides a bidirectional secret forwarding mechanism
// between two providers, with optional key transformation and filtering.
package bridge

import (
	"context"
	"errors"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Direction controls which way secrets flow.
type Direction int

const (
	DirectionAtoB Direction = iota
	DirectionBtoA
	DirectionBoth
)

// Result holds the outcome of a single forwarded key.
type Result struct {
	Key   string
	Error error
}

// Bridge forwards secrets between two providers.
type Bridge struct {
	reg       *provider.Registry
	srcName   string
	dstName   string
	transform func(string) string
	filter    func(string) bool
}

// New creates a Bridge between srcName and dstName providers.
func New(reg *provider.Registry, srcName, dstName string) (*Bridge, error) {
	if reg == nil {
		return nil, errors.New("bridge: registry must not be nil")
	}
	if srcName == "" {
		return nil, errors.New("bridge: source provider name must not be empty")
	}
	if dstName == "" {
		return nil, errors.New("bridge: destination provider name must not be empty")
	}
	return &Bridge{
		reg:     reg,
		srcName: srcName,
		dstName: dstName,
	}, nil
}

// WithTransform sets a key transformation function applied before writing.
func (b *Bridge) WithTransform(fn func(string) string) *Bridge {
	b.transform = fn
	return b
}

// WithFilter sets a predicate; only keys for which fn returns true are forwarded.
func (b *Bridge) WithFilter(fn func(string) bool) *Bridge {
	b.filter = fn
	return b
}

// Forward copies all secrets from the source to the destination provider.
func (b *Bridge) Forward(ctx context.Context) ([]Result, error) {
	src, ok := b.reg.Get(b.srcName)
	if !ok {
		return nil, fmt.Errorf("bridge: unknown source provider %q", b.srcName)
	}
	dst, ok := b.reg.Get(b.dstName)
	if !ok {
		return nil, fmt.Errorf("bridge: unknown destination provider %q", b.dstName)
	}

	keys, err := src.ListSecrets(ctx)
	if err != nil {
		return nil, fmt.Errorf("bridge: list source secrets: %w", err)
	}

	var results []Result
	for _, key := range keys {
		if b.filter != nil && !b.filter(key) {
			continue
		}
		val, err := src.GetSecret(ctx, key)
		if err != nil {
			results = append(results, Result{Key: key, Error: err})
			continue
		}
		dstKey := key
		if b.transform != nil {
			dstKey = b.transform(key)
		}
		if err := dst.PutSecret(ctx, dstKey, val); err != nil {
			results = append(results, Result{Key: key, Error: err})
			continue
		}
		results = append(results, Result{Key: key})
	}
	return results, nil
}

// HasFailures returns true if any result contains an error.
func HasFailures(results []Result) bool {
	for _, r := range results {
		if r.Error != nil {
			return true
		}
	}
	return false
}
