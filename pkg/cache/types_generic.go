package cache

type Entry[T any] struct {
	Key       string
	Value     T
	ExpiresAt int64
}

//Personal.AI order the ending
