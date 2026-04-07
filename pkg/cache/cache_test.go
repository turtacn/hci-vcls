package cache

import (
	"sort"
	"testing"
	"time"
)

func TestCache_SetGet(t *testing.T) {
	c := NewMemoryCache[string, string](1 * time.Minute)
	defer c.Close()

	c.Set("key1", "val1", 0)

	val, found := c.Get("key1")
	if !found || val != "val1" {
		t.Errorf("Expected val1, got %v (found: %v)", val, found)
	}

	_, found = c.Get("key2")
	if found {
		t.Errorf("Expected not found for key2")
	}
}

func TestCache_TTL(t *testing.T) {
	c := NewMemoryCache[string, string](10 * time.Millisecond)
	defer c.Close()

	c.Set("key1", "val1", 50*time.Millisecond)

	val, found := c.Get("key1")
	if !found || val != "val1" {
		t.Errorf("Expected val1 immediately, got %v", val)
	}

	time.Sleep(100 * time.Millisecond)

	_, found = c.Get("key1")
	if found {
		t.Errorf("Expected key1 to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	c := NewMemoryCache[string, string](1 * time.Minute)
	defer c.Close()

	c.Set("key1", "val1", 0)
	c.Delete("key1")

	_, found := c.Get("key1")
	if found {
		t.Errorf("Expected key1 to be deleted")
	}
}

func TestCache_Keys(t *testing.T) {
	c := NewMemoryCache[string, string](1 * time.Minute)
	defer c.Close()

	c.Set("key1", "val1", 0)
	c.Set("key2", "val2", 0)
	c.Set("key3", "val3", 10*time.Millisecond)

	time.Sleep(50 * time.Millisecond)

	keys := c.Keys()
	sort.Strings(keys)

	if len(keys) != 2 || keys[0] != "key1" || keys[1] != "key2" {
		t.Errorf("Expected [key1 key2], got %v", keys)
	}
}

