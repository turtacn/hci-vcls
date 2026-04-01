package cache

import (
	"sync"
	"time"
)

type item[V any] struct {
	value      V
	expiration int64
}

type MemoryCache[K comparable, V any] struct {
	mu     sync.RWMutex
	items  map[K]item[V]
	closed chan struct{}
}

var _ Cache[string, any] = &MemoryCache[string, any]{}

func NewMemoryCache[K comparable, V any](cleanupInterval time.Duration) *MemoryCache[K, V] {
	c := &MemoryCache[K, V]{
		items:  make(map[K]item[V]),
		closed: make(chan struct{}),
	}
	go c.cleanupLoop(cleanupInterval)
	return c
}

func (c *MemoryCache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	c.items[key] = item[V]{
		value:      value,
		expiration: exp,
	}
}

func (c *MemoryCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		var zero V
		return zero, false
	}

	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		var zero V
		return zero, false
	}

	return item.value, true
}

func (c *MemoryCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

func (c *MemoryCache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.items))
	now := time.Now().UnixNano()

	for k, v := range c.items {
		if v.expiration == 0 || now <= v.expiration {
			keys = append(keys, k)
		}
	}
	return keys
}

func (c *MemoryCache[K, V]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]item[V])
}

func (c *MemoryCache[K, V]) Close() error {
	select {
	case <-c.closed:
	default:
		close(c.closed)
	}
	return nil
}

func (c *MemoryCache[K, V]) cleanupLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.closed:
			return
		case <-ticker.C:
			c.cleanup()
		}
	}
}

func (c *MemoryCache[K, V]) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now().UnixNano()
	for k, v := range c.items {
		if v.expiration > 0 && now > v.expiration {
			delete(c.items, k)
		}
	}
}

// Personal.AI order the ending
