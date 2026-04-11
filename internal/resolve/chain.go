package resolve

import (
	"context"
	"fmt"
)

// Chain is an ordered list of named providers used for transparent fallback.
type Chain struct {
	entries []chainEntry
}

type chainEntry struct {
	name     string
	provider Provider
}

// NewChain constructs an empty Chain.
func NewChain() *Chain {
	return &Chain{}
}

// Add appends a named provider to the end of the chain.
func (c *Chain) Add(name string, p Provider) {
	c.entries = append(c.entries, chainEntry{name: name, provider: p})
}

// Get walks the chain in registration order, returning the first successful
// result along with the name of the provider that satisfied the request.
func (c *Chain) Get(ctx context.Context, key string) (value, source string, err error) {
	for _, e := range c.entries {
		v, gerr := e.provider.Get(ctx, key)
		if gerr == nil {
			return v, e.name, nil
		}
	}
	return "", "", fmt.Errorf("%w: %s", ErrNotResolved, key)
}

// Providers returns a plain []Provider slice suitable for use with New.
func (c *Chain) Providers() []Provider {
	out := make([]Provider, len(c.entries))
	for i, e := range c.entries {
		out[i] = e.provider
	}
	return out
}
