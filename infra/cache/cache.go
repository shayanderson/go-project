package cache

import (
	"sync"
	"sync/atomic"
)

// Cache is a simple in-memory cache with metrics
type Cache[T any, K comparable] struct {
	*metrics
	mu    sync.RWMutex
	store map[K]T
}

// New creates a new Cache instance
func New[T any, K comparable]() *Cache[T, K] {
	return &Cache[T, K]{
		metrics: &metrics{},
		store:   make(map[K]T),
	}
}

// Clear clears all items from the cache
func (c *Cache[T, K]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store = make(map[K]T)
}

// Delete deletes an item from the cache by key
func (c *Cache[T, K]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.store, key)
}

// Get retrieves an item from the cache by key
func (c *Cache[T, K]) Get(key K) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	v, ok := c.store[key]
	if ok {
		c.hits.Add(1)
	} else {
		c.misses.Add(1)
	}
	return v, ok
}

// Put adds an item to the cache
func (c *Cache[T, K]) Put(key K, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.store[key] = value
}

// Size returns the number of items in the cache
func (c *Cache[T, K]) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.store)
}

// Metrics holds cache metrics data
type Metrics struct {
	Hits   int64
	Misses int64
}

// metrics holds internal metrics data
type metrics struct {
	hits   atomic.Int64
	misses atomic.Int64
}

// Metrics returns the current cache metrics
func (m *metrics) Metrics() Metrics {
	return Metrics{
		Hits:   m.hits.Load(),
		Misses: m.misses.Load(),
	}
}
