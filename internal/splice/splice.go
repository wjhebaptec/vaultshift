// Package splice merges secrets from multiple source providers into a single
// destination provider, applying an optional key-rewrite function before writing.
package splice

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Result holds the outcome of a single splice operation.
type Result struct {
	SourceProvider string
	Key            string
	DestKey        string
	Err            error
}

// Splicer copies secrets from many sources into one destination.
type Splicer struct {
	reg     *provider.Registry
	dest    string
	rewrite func(source, key string) string
	results []Result
}

// New creates a Splicer that writes into destProvider.
// rewrite may be nil; if set it transforms (sourceProvider, key) -> destKey.
func New(reg *provider.Registry, destProvider string, rewrite func(source, key string) string) (*Splicer, error) {
	if reg == nil {
		return nil, fmt.Errorf("splice: registry must not be nil")
	}
	if destProvider == "" {
		return nil, fmt.Errorf("splice: destination provider must not be empty")
	}
	return &Splicer{reg: reg, dest: destProvider, rewrite: rewrite}, nil
}

// Splice reads key from srcProvider and writes it to the destination.
func (s *Splicer) Splice(ctx context.Context, srcProvider, key string) error {
	src, err := s.reg.Get(srcProvider)
	if err != nil {
		return fmt.Errorf("splice: unknown source provider %q: %w", srcProvider, err)
	}
	dst, err := s.reg.Get(s.dest)
	if err != nil {
		return fmt.Errorf("splice: unknown destination provider %q: %w", s.dest, err)
	}
	val, err := src.GetSecret(ctx, key)
	if err != nil {
		return fmt.Errorf("splice: get %q from %q: %w", key, srcProvider, err)
	}
	destKey := key
	if s.rewrite != nil {
		destKey = s.rewrite(srcProvider, key)
	}
	if err := dst.PutSecret(ctx, destKey, val); err != nil {
		return fmt.Errorf("splice: put %q into %q: %w", destKey, s.dest, err)
	}
	s.results = append(s.results, Result{SourceProvider: srcProvider, Key: key, DestKey: destKey})
	return nil
}

// SpliceAll reads every key listed in keys from srcProvider and writes each to dest.
func (s *Splicer) SpliceAll(ctx context.Context, srcProvider string, keys []string) {
	for _, k := range keys {
		err := s.Splice(ctx, srcProvider, k)
		if err != nil {
			s.results = append(s.results, Result{SourceProvider: srcProvider, Key: k, Err: err})
		}
	}
}

// Results returns the recorded outcomes.
func (s *Splicer) Results() []Result { return s.results }

// HasFailures returns true if any splice operation recorded an error.
func (s *Splicer) HasFailures() bool {
	for _, r := range s.results {
		if r.Err != nil {
			return true
		}
	}
	return false
}
