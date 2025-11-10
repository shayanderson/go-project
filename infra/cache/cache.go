package cache

import (
	"sync"
)

// Cache is a simple in-memory cache
type Cache[T any, K comparable] struct {
	mu    sync.RWMutex
	store map[K]T
}

// New creates a new Cache instance
func New[T any, K comparable]() *Cache[T, K] {
	return &Cache[T, K]{
		store: make(map[K]T),
	}
}

// All returns all items in the cache
func (c *Cache[T, K]) All() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	r := make([]T, 0, len(c.store))
	for _, v := range c.store {
		r = append(r, v)
	}
	return r
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
