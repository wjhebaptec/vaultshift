// Package fence provides a monotonic write fence that rejects out-of-order
// or duplicate writes based on a per-key sequence number.
package fence

import (
	"errors"
	"fmt"
	"sync"
)

// ErrOutOfOrder is returned when the provided sequence is not strictly greater
// than the last accepted sequence for the key.
var ErrOutOfOrder = errors.New("fence: out-of-order sequence")

// ErrUnknownKey is returned when querying a key that has never been seen.
var ErrUnknownKey = errors.New("fence: unknown key")

// Fence tracks the latest accepted sequence number per key.
type Fence struct {
	mu      sync.Mutex
	seqs    map[string]uint64
}

// New creates a new Fence.
func New() *Fence {
	return &Fence{seqs: make(map[string]uint64)}
}

// Check returns nil if seq is strictly greater than the last accepted sequence
// for key, and records it. Otherwise it returns ErrOutOfOrder.
func (f *Fence) Check(key string, seq uint64) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if last, ok := f.seqs[key]; ok && seq <= last {
		return fmt.Errorf("%w: key=%q got=%d last=%d", ErrOutOfOrder, key, seq, last)
	}
	f.seqs[key] = seq
	return nil
}

// Latest returns the last accepted sequence for key.
func (f *Fence) Latest(key string) (uint64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	v, ok := f.seqs[key]
	if !ok {
		return 0, fmt.Errorf("%w: %q", ErrUnknownKey, key)
	}
	return v, nil
}

// Reset clears the sequence record for key.
func (f *Fence) Reset(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.seqs, key)
}

// ResetAll clears all sequence records.
func (f *Fence) ResetAll() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.seqs = make(map[string]uint64)
}
