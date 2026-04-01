package cache

import (
	"testing"
	"time"
)

func TestMemoryCache_SetGet(t *testing.T) {
	cache := NewMemoryCache[string]()
	key := "test_key"
	val := "test_val"

	err := cache.Set(key, val, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error setting cache: %v", err)
	}

	got, err := cache.Get(key)
	if err != nil {
		t.Fatalf("unexpected error getting cache: %v", err)
	}
	if got != val {
		t.Fatalf("expected %v, got %v", val, got)
	}
}

func TestMemoryCache_Expired(t *testing.T) {
	cache := NewMemoryCache[int]()
	key := "test_key"
	val := 42

	err := cache.Set(key, val, time.Millisecond*10)
	if err != nil {
		t.Fatalf("unexpected error setting cache: %v", err)
	}

	time.Sleep(time.Millisecond * 20)
	_, err = cache.Get(key)
	if err != ErrCacheMiss {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache[bool]()
	key := "test_key"

	cache.Set(key, true, time.Minute)
	cache.Delete(key)

	_, err := cache.Get(key)
	if err != ErrCacheMiss {
		t.Fatalf("expected ErrCacheMiss, got %v", err)
	}
}

func TestMemoryCache_Keys(t *testing.T) {
	cache := NewMemoryCache[string]()
	cache.Set("k1", "v1", time.Minute)
	cache.Set("k2", "v2", time.Minute)
	cache.Set("k3", "v3", time.Millisecond*10)

	time.Sleep(time.Millisecond * 20) // Let k3 expire

	keys := cache.Keys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

//Personal.AI order the ending
