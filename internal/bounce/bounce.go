// Package bounce provides a write-back buffer that accumulates Put operations
// and flushes them to the underlying provider in a single batch when the buffer
// is full or when Flush is called explicitly.
package bounce

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/vaultshift/internal/provider"
)

// ErrFlushing is returned when a flush fails for one or more keys.
var ErrFlushing = errors.New("bounce: flush error")

// FlushResult holds the outcome of a single buffered write.
type FlushResult struct {
	Key   string
	Error error
}

// Buffer accumulates writes and flushes them in batches.
type Buffer struct {
	mu       sync.Mutex
	buf      map[string]string
	batchSz  int
	provider provider.Provider
}

// New creates a Buffer that flushes automatically once batchSize pending
// writes have accumulated. batchSize must be >= 1.
func New(p provider.Provider, batchSize int) (*Buffer, error) {
	if p == nil {
		return nil, errors.New("bounce: provider must not be nil")
	}
	if batchSize < 1 {
		return nil, errors.New("bounce: batchSize must be >= 1")
	}
	return &Buffer{
		buf:     make(map[string]string),
		batchSz: batchSize,
		provider: p,
	}, nil
}

// Put buffers a key/value pair. If the buffer reaches its batch size the
// contents are flushed automatically.
func (b *Buffer) Put(ctx context.Context, key, value string) error {
	b.mu.Lock()
	b.buf[key] = value
	ready := len(b.buf) >= b.batchSz
	b.mu.Unlock()

	if ready {
		results := b.Flush(ctx)
		for _, r := range results {
			if r.Error != nil {
				return fmt.Errorf("%w: key %q: %v", ErrFlushing, r.Key, r.Error)
			}
		}
	}
	return nil
}

// Flush writes all buffered entries to the underlying provider and clears the
// buffer. It returns one FlushResult per entry.
func (b *Buffer) Flush(ctx context.Context) []FlushResult {
	b.mu.Lock()
	snapshot := make(map[string]string, len(b.buf))
	for k, v := range b.buf {
		snapshot[k] = v
	}
	b.buf = make(map[string]string)
	b.mu.Unlock()

	results := make([]FlushResult, 0, len(snapshot))
	for k, v := range snapshot {
		err := b.provider.Put(ctx, k, v)
		results = append(results, FlushResult{Key: k, Error: err})
	}
	return results
}

// Len returns the number of entries currently held in the buffer.
func (b *Buffer) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.buf)
}
