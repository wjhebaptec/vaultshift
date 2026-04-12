// Package promote handles promotion of secrets from one environment to another
// (e.g. staging → production), optionally transforming keys in the process.
package promote

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Options configures a promotion run.
type Options struct {
	// KeyTransform is an optional function applied to each key before writing
	// to the destination. If nil the key is used as-is.
	KeyTransform func(key string) string

	// DryRun causes Promote / PromoteAll to report what would happen without
	// writing anything to the destination provider.
	DryRun bool
}

// Result captures the outcome of a single key promotion.
type Result struct {
	SourceKey string
	DestKey   string
	DryRun    bool
	Err       error
}

// Promoter copies secrets from a source provider to a destination provider.
type Promoter struct {
	reg  *provider.Registry
	opts Options
}

// New creates a Promoter that uses reg to look up providers.
func New(reg *provider.Registry, opts Options) *Promoter {
	return &Promoter{reg: reg, opts: opts}
}

// Promote copies a single key from srcProvider to dstProvider.
func (p *Promoter) Promote(ctx context.Context, srcProvider, dstProvider, key string) Result {
	src, ok := p.reg.Get(srcProvider)
	if !ok {
		return Result{SourceKey: key, Err: fmt.Errorf("promote: unknown source provider %q", srcProvider)}
	}
	dst, ok := p.reg.Get(dstProvider)
	if !ok {
		return Result{SourceKey: key, Err: fmt.Errorf("promote: unknown destination provider %q", dstProvider)}
	}

	value, err := src.Get(ctx, key)
	if err != nil {
		return Result{SourceKey: key, Err: fmt.Errorf("promote: get %q from %q: %w", key, srcProvider, err)}
	}

	destKey := key
	if p.opts.KeyTransform != nil {
		destKey = p.opts.KeyTransform(key)
	}

	if p.opts.DryRun {
		return Result{SourceKey: key, DestKey: destKey, DryRun: true}
	}

	if err := dst.Put(ctx, destKey, value); err != nil {
		return Result{SourceKey: key, DestKey: destKey, Err: fmt.Errorf("promote: put %q into %q: %w", destKey, dstProvider, err)}
	}

	return Result{SourceKey: key, DestKey: destKey}
}

// PromoteAll promotes every key returned by srcProvider.List to dstProvider.
func (p *Promoter) PromoteAll(ctx context.Context, srcProvider, dstProvider string) []Result {
	src, ok := p.reg.Get(srcProvider)
	if !ok {
		return []Result{{Err: fmt.Errorf("promote: unknown source provider %q", srcProvider)}}
	}

	keys, err := src.List(ctx)
	if err != nil {
		return []Result{{Err: fmt.Errorf("promote: list from %q: %w", srcProvider, err)}}
	}

	results := make([]Result, 0, len(keys))
	for _, k := range keys {
		results = append(results, p.Promote(ctx, srcProvider, dstProvider, k))
	}
	return results
}

// HasFailures returns true if any result in rs contains a non-nil error.
func HasFailures(rs []Result) bool {
	for _, r := range rs {
		if r.Err != nil {
			return true
		}
	}
	return false
}
