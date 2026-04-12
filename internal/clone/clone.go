// Package clone provides functionality for duplicating secrets from one
// provider and key to another, optionally transforming the key name.
package clone

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Options configures a Clone operation.
type Options struct {
	// KeyTransform is an optional function to rename the key in the destination.
	// If nil, the original key is used.
	KeyTransform func(key string) string

	// DryRun prevents writes when true.
	DryRun bool
}

// Result holds the outcome of a single clone operation.
type Result struct {
	SourceKey string
	DestKey   string
	Provider  string
	Skipped   bool
	Err       error
}

// Cloner copies secrets between providers.
type Cloner struct {
	registry *provider.Registry
}

// New creates a new Cloner backed by the given registry.
func New(registry *provider.Registry) *Cloner {
	return &Cloner{registry: registry}
}

// Clone copies a secret from srcProvider/srcKey to dstProvider/dstKey.
// If opts.KeyTransform is set, the destination key is derived from srcKey.
func (c *Cloner) Clone(ctx context.Context, srcProvider, dstProvider, srcKey string, opts Options) Result {
	destKey := srcKey
	if opts.KeyTransform != nil {
		destKey = opts.KeyTransform(srcKey)
	}

	res := Result{SourceKey: srcKey, DestKey: destKey, Provider: dstProvider}

	src, err := c.registry.Get(srcProvider)
	if err != nil {
		res.Err = fmt.Errorf("clone: source provider %q: %w", srcProvider, err)
		return res
	}

	dst, err := c.registry.Get(dstProvider)
	if err != nil {
		res.Err = fmt.Errorf("clone: destination provider %q: %w", dstProvider, err)
		return res
	}

	value, err := src.GetSecret(ctx, srcKey)
	if err != nil {
		res.Err = fmt.Errorf("clone: read %q from %q: %w", srcKey, srcProvider, err)
		return res
	}

	if opts.DryRun {
		res.Skipped = true
		return res
	}

	if err := dst.PutSecret(ctx, destKey, value); err != nil {
		res.Err = fmt.Errorf("clone: write %q to %q: %w", destKey, dstProvider, err)
		return res
	}

	return res
}

// CloneAll clones every key returned by srcProvider.ListSecrets into dstProvider.
func (c *Cloner) CloneAll(ctx context.Context, srcProvider, dstProvider string, opts Options) []Result {
	src, err := c.registry.Get(srcProvider)
	if err != nil {
		return []Result{{Err: fmt.Errorf("clone: source provider %q: %w", srcProvider, err)}}
	}

	keys, err := src.ListSecrets(ctx)
	if err != nil {
		return []Result{{Err: fmt.Errorf("clone: list secrets from %q: %w", srcProvider, err)}}
	}

	results := make([]Result, 0, len(keys))
	for _, key := range keys {
		results = append(results, c.Clone(ctx, srcProvider, dstProvider, key, opts))
	}
	return results
}
