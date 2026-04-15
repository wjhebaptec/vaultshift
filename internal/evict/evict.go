// Package evict provides a least-recently-used (LRU) eviction layer over a
// secret provider. When the cache reaches its capacity the oldest entry is
// removed before a new one is inserted.
package evict

import (
	"container/list"
	"context"
	"errors"
	"sync"

	"github.com/vaultshift/internal/provider"
)

// ErrInvalidCapacity is returned when capacity is less than 1.
var ErrInvalidCapacity = errors.New("evict: capacity must be at least 1")

type entry struct {
	key   string
	value string
}

// Cache is an LRU-evicting wrapper around a provider.
type Cache struct {
	mu       sync.Mutex
	cap      int
	list     *list.List
	items    map[string]*list.Element
	backend  provider.Provider
}

// New creates a new LRU Cache with the given capacity backed by p.
func New(p provider.Provider, capacity int) (*Cache, error) {
	if p == nil {
		return nil, errors.New("evict: provider must not be nil")
	}
	if capacity < 1 {
		return nil, ErrInvalidCapacity
	}
	return &Cache{
		cap:     capacity,
		list:    list.New(),
		items:   make(map[string]*list.Element),
		backend: p,
	}, nil
}

// Put stores the secret in the backend and updates the LRU cache.
func (c *Cache) Put(ctx context.Context, key, value string) error {
	if err := c.backend.Put(ctx, key, value); err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.promote(key, value)
	return nil
}

// Get retrieves a secret, consulting the LRU cache before the backend.
func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	c.mu.Lock()
	if el, ok := c.items[key]; ok {
		c.list.MoveToFront(el)
		v := el.Value.(*entry).value
		c.mu.Unlock()
		return v, nil
	}
	c.mu.Unlock()

	val, err := c.backend.Get(ctx, key)
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	c.promote(key, val)
	c.mu.Unlock()
	return val, nil
}

// Delete removes the secret from both the cache and the backend.
func (c *Cache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	if el, ok := c.items[key]; ok {
		c.list.Remove(el)
		delete(c.items, key)
	}
	c.mu.Unlock()
	return c.backend.Delete(ctx, key)
}

// List delegates directly to the backend.
func (c *Cache) List(ctx context.Context) ([]string, error) {
	return c.backend.List(ctx)
}

// Len returns the number of entries currently held in the LRU cache.
func (c *Cache) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.list.Len()
}

// promote inserts or refreshes key/value and evicts the LRU entry if needed.
// Must be called with c.mu held.
func (c *Cache) promote(key, value string) {
	if el, ok := c.items[key]; ok {
		el.Value.(*entry).value = value
		c.list.MoveToFront(el)
		return
	}
	if c.list.Len() >= c.cap {
		oldest := c.list.Back()
		if oldest != nil {
			c.list.Remove(oldest)
			delete(c.items, oldest.Value.(*entry).key)
		}
	}
	el := c.list.PushFront(&entry{key: key, value: value})
	c.items[key] = el
}
