// Package bloom provides a probabilistic membership filter for secret keys,
// allowing fast existence checks without querying the underlying provider.
package bloom

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"sync"
)

const defaultSize = 1024

// Filter is a thread-safe bloom filter for secret key membership testing.
type Filter struct {
	mu   sync.RWMutex
	bits []bool
	k    int // number of hash functions
	n    int // number of bits
}

// Option configures a Filter.
type Option func(*Filter)

// WithSize sets the number of bits in the filter (must be > 0).
func WithSize(n int) Option {
	return func(f *Filter) { f.n = n }
}

// WithHashFunctions sets the number of independent hash functions.
func WithHashFunctions(k int) Option {
	return func(f *Filter) { f.k = k }
}

// New creates a new bloom Filter with optional configuration.
func New(opts ...Option) (*Filter, error) {
	f := &Filter{n: defaultSize, k: 3}
	for _, o := range opts {
		o(f)
	}
	if f.n <= 0 {
		return nil, errors.New("bloom: size must be greater than zero")
	}
	if f.k <= 0 {
		return nil, errors.New("bloom: hash function count must be greater than zero")
	}
	f.bits = make([]bool, f.n)
	return f, nil
}

// Add inserts a key into the filter.
func (f *Filter) Add(key string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, idx := range f.indices(key) {
		f.bits[idx] = true
	}
}

// MayContain returns true if the key was possibly added, false if definitely not.
func (f *Filter) MayContain(key string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, idx := range f.indices(key) {
		if !f.bits[idx] {
			return false
		}
	}
	return true
}

// Reset clears all bits in the filter.
func (f *Filter) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range f.bits {
		f.bits[i] = false
	}
}

// indices returns k bit positions for the given key.
func (f *Filter) indices(key string) []int {
	idxs := make([]int, f.k)
	for i := 0; i < f.k; i++ {
		h := sha256.Sum256([]byte(key + string(rune(i))))
		v := binary.BigEndian.Uint64(h[:8])
		idxs[i] = int(v % uint64(f.n))
	}
	return idxs
}
