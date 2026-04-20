// Package shred provides secure deletion of secrets from one or more providers.
// It overwrites the secret value before deleting it, reducing the risk of
// recovery from provider-level storage snapshots.
package shred

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// Provider is the minimal interface required by the shredder.
type Provider interface {
	Put(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}

// Result holds the outcome of a single shred operation.
type Result struct {
	Key      string
	Provider string
	Err      error
}

// Shredder securely deletes secrets by overwriting then deleting them.
type Shredder struct {
	providers map[string]Provider
	passes    int
}

// Option configures a Shredder.
type Option func(*Shredder)

// WithPasses sets the number of overwrite passes before deletion.
func WithPasses(n int) Option {
	return func(s *Shredder) {
		if n > 0 {
			s.passes = n
		}
	}
}

// New creates a Shredder for the given named providers.
func New(providers map[string]Provider, opts ...Option) (*Shredder, error) {
	if len(providers) == 0 {
		return nil, errors.New("shred: at least one provider is required")
	}
	s := &Shredder{providers: providers, passes: 1}
	for _, o := range opts {
		o(s)
	}
	return s, nil
}

// Shred overwrites key with random data for each configured pass, then deletes it
// from the named provider.
func (s *Shredder) Shred(ctx context.Context, providerName, key string) error {
	p, ok := s.providers[providerName]
	if !ok {
		return fmt.Errorf("shred: unknown provider %q", providerName)
	}
	for i := 0; i < s.passes; i++ {
		rnd, err := randomValue(32)
		if err != nil {
			return fmt.Errorf("shred: generating random overwrite value: %w", err)
		}
		if err := p.Put(ctx, key, rnd); err != nil {
			return fmt.Errorf("shred: overwrite pass %d for key %q: %w", i+1, key, err)
		}
	}
	if err := p.Delete(ctx, key); err != nil {
		return fmt.Errorf("shred: delete key %q: %w", key, err)
	}
	return nil
}

// ShredAll shreds the given key from every registered provider and returns all results.
func (s *Shredder) ShredAll(ctx context.Context, key string) []Result {
	results := make([]Result, 0, len(s.providers))
	for name := range s.providers {
		err := s.Shred(ctx, name, key)
		results = append(results, Result{Key: key, Provider: name, Err: err})
	}
	return results
}

// HasFailures reports whether any result contains an error.
func HasFailures(results []Result) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}

func randomValue(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf), nil
}
