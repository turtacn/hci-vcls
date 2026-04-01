package cache

import "time"

type Cache[T any] interface {
	Get(key string) (T, error)
	Set(key string, value T, ttl time.Duration) error
	Delete(key string) error
	Keys() []string
}

//Personal.AI order the ending
