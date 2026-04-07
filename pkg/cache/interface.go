package cache

import "time"

type Cache[K comparable, V any] interface {
	Set(key K, value V, ttl time.Duration)
	Get(key K) (V, bool)
	Delete(key K)
	Keys() []K
	Flush()
	Close() error
}

