// Package swap provides atomic secret swapping between two providers.
// It reads the current value from a source, writes it to a destination,
// and optionally writes the destination's old value back to the source.
package swap

import (
	"context"
	"errors"
	"fmt"

	"github.com/vaultshift/internal/provider"
)

// Result holds the outcome of a single swap operation.
type Result struct {
	Key      string
	OldSrc   string
	OldDst   string
	Swapped  bool
	Err      error
}

// Swapper exchanges secrets between two registered providers.
type Swapper struct {
	reg      *provider.Registry
	bidirect bool
}

// New creates a Swapper. When bidirectional is true the destination's
// previous value is written back to the source.
func New(reg *provider.Registry, bidirectional bool) (*Swapper, error) {
	if reg == nil {
		return nil, errors.New("swap: registry must not be nil")
	}
	return &Swapper{reg: reg, bidirect: bidirectional}, nil
}

// Swap exchanges the secret identified by key between srcName and dstName.
func (s *Swapper) Swap(ctx context.Context, srcName, dstName, key string) Result {
	res := Result{Key: key}

	src, ok := s.reg.Get(srcName)
	if !ok {
		res.Err = fmt.Errorf("swap: unknown source provider %q", srcName)
		return res
	}
	dst, ok := s.reg.Get(dstName)
	if !ok {
		res.Err = fmt.Errorf("swap: unknown destination provider %q", dstName)
		return res
	}

	srcVal, err := src.Get(ctx, key)
	if err != nil {
		res.Err = fmt.Errorf("swap: read source: %w", err)
		return res
	}
	res.OldSrc = srcVal

	dstVal, _ := dst.Get(ctx, key) // best-effort; may not exist
	res.OldDst = dstVal

	if err := dst.Put(ctx, key, srcVal); err != nil {
		res.Err = fmt.Errorf("swap: write destination: %w", err)
		return res
	}

	if s.bidirect && dstVal != "" {
		if err := src.Put(ctx, key, dstVal); err != nil {
			res.Err = fmt.Errorf("swap: write source (bidirectional): %w", err)
			return res
		}
	}

	res.Swapped = true
	return res
}

// SwapAll swaps every key in keys and returns all results.
func (s *Swapper) SwapAll(ctx context.Context, srcName, dstName string, keys []string) []Result {
	out := make([]Result, 0, len(keys))
	for _, k := range keys {
		out = append(out, s.Swap(ctx, srcName, dstName, k))
	}
	return out
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
