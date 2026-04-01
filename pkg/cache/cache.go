package cache

import (
	"sync"
	"time"
)

type memoryCache[T any] struct {
	mu    sync.RWMutex
	store map[string]Entry[T]
}

func NewMemoryCache[T any]() Cache[T] {
	return &memoryCache[T]{
		store: make(map[string]Entry[T]),
	}
}

func (c *memoryCache[T]) Get(key string) (T, error) {
	c.mu.RLock()
	entry, ok := c.store[key]
	c.mu.RUnlock()

	var zero T
	if !ok {
		return zero, ErrCacheMiss
	}

	if entry.ExpiresAt > 0 && time.Now().UnixNano() > entry.ExpiresAt {
		c.Delete(key)
		return zero, ErrCacheMiss
	}

	return entry.Value, nil
}

func (c *memoryCache[T]) Set(key string, value T, ttl time.Duration) error {
	var expiresAt int64
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	c.store[key] = Entry[T]{
		Key:       key,
		Value:     value,
		ExpiresAt: expiresAt,
	}
	c.mu.Unlock()

	return nil
}

func (c *memoryCache[T]) Delete(key string) error {
	c.mu.Lock()
	delete(c.store, key)
	c.mu.Unlock()
	return nil
}

func (c *memoryCache[T]) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0, len(c.store))
	now := time.Now().UnixNano()
	for k, v := range c.store {
		if v.ExpiresAt == 0 || now <= v.ExpiresAt {
			keys = append(keys, k)
		}
	}
	return keys
}

//Personal.AI order the ending
