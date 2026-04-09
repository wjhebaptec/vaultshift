// Package cache provides an in-memory TTL cache for secret values,
// reducing redundant reads from remote secret managers.
package cache

import (
	"sync"
	"time"
)

// Entry holds a cached secret value along with its expiry time.
type Entry struct {
	Value     string
	ExpiresAt time.Time
}

// Cache is a thread-safe, TTL-based in-memory store for secrets.
type Cache struct {
	mu      sync.RWMutex
	items   map[string]Entry
	ttl     time.Duration
	nowFunc func() time.Time
}

// New creates a new Cache with the given TTL duration.
// A zero TTL means entries never expire.
func New(ttl time.Duration) *Cache {
	return &Cache{
		items:   make(map[string]Entry),
		ttl:     ttl,
		nowFunc: time.Now,
	}
}

// Set stores a secret value under the given key.
func (c *Cache) Set(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if c.ttl > 0 {
		expiresAt = c.nowFunc().Add(c.ttl)
	}
	c.items[key] = Entry{Value: value, ExpiresAt: expiresAt}
}

// Get retrieves a secret value by key. Returns the value and true if found
// and not expired, otherwise returns an empty string and false.
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.items[key]
	if !ok {
		return "", false
	}
	if !entry.ExpiresAt.IsZero() && c.nowFunc().After(entry.ExpiresAt) {
		return "", false
	}
	return entry.Value, true
}

// Delete removes a key from the cache.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Flush removes all entries from the cache.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]Entry)
}

// Size returns the number of entries currently in the cache,
// including those that may have expired but not yet been evicted.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
