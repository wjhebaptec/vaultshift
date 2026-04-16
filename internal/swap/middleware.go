package swap

import (
	"context"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// GuardedSwapper wraps a Swapper and enforces that both providers are
// reachable before attempting any swap.
type GuardedSwapper struct {
	inner *Swapper
	reg   *provider.Registry
}

// NewGuarded returns a GuardedSwapper that pre-validates provider names.
func NewGuarded(reg *provider.Registry, bidirectional bool) (*GuardedSwapper, error) {
	s, err := New(reg, bidirectional)
	if err != nil {
		return nil, err
	}
	return &GuardedSwapper{inner: s, reg: reg}, nil
}

// Swap validates provider availability then delegates to the inner Swapper.
func (g *GuardedSwapper) Swap(ctx context.Context, srcName, dstName, key string) Result {
	if _, ok := g.reg.Get(srcName); !ok {
		return Result{Key: key, Err: fmt.Errorf("guarded swap: source provider %q not registered", srcName)}
	}
	if _, ok := g.reg.Get(dstName); !ok {
		return Result{Key: key, Err: fmt.Errorf("guarded swap: destination provider %q not registered", dstName)}
	}
	return g.inner.Swap(ctx, srcName, dstName, key)
}

// SwapAll validates providers once then swaps all keys.
func (g *GuardedSwapper) SwapAll(ctx context.Context, srcName, dstName string, keys []string) []Result {
	if _, ok := g.reg.Get(srcName); !ok {
		out := make([]Result, len(keys))
		for i, k := range keys {
			out[i] = Result{Key: k, Err: fmt.Errorf("guarded swap: source provider %q not registered", srcName)}
		}
		return out
	}
	if _, ok := g.reg.Get(dstName); !ok {
		out := make([]Result, len(keys))
		for i, k := range keys {
			out[i] = Result{Key: k, Err: fmt.Errorf("guarded swap: destination provider %q not registered", dstName)}
		}
		return out
	}
	return g.inner.SwapAll(ctx, srcName, dstName, keys)
}
