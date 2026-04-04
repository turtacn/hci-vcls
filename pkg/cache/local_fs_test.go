package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLocalStore_GetPutDelete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStore(dir)
	if err != nil {
		t.Fatalf("Failed to create local store: %v", err)
	}

	meta := VMComputeMeta{VMID: "100", CPUs: 4, Memory: 8192, NodeID: "node-1", GroupID: "group-a"}

	// Get before put should miss
	_, err = store.Get("100")
	if err != ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss, got %v", err)
	}

	// Put
	err = store.Put("100", meta)
	if err != nil {
		t.Fatalf("Failed to put meta: %v", err)
	}

	// Get after put
	gotMeta, err := store.Get("100")
	if err != nil {
		t.Fatalf("Failed to get meta: %v", err)
	}
	if gotMeta.CPUs != meta.CPUs || gotMeta.Memory != meta.Memory || gotMeta.NodeID != meta.NodeID {
		t.Errorf("Retrieved meta does not match. Expected %v, got %v", meta, *gotMeta)
	}

	// Delete
	err = store.Delete("100")
	if err != nil {
		t.Fatalf("Failed to delete meta: %v", err)
	}

	// Get after delete should miss
	_, err = store.Get("100")
	if err != ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss, got %v", err)
	}
}

func TestLocalStore_List(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStore(dir)
	if err != nil {
		t.Fatalf("Failed to create local store: %v", err)
	}

	meta1 := VMComputeMeta{VMID: "100"}
	meta2 := VMComputeMeta{VMID: "101"}

	_ = store.Put("100", meta1)
	_ = store.Put("101", meta2)

	// Add an invalid file to ensure it doesn't break
	_ = os.WriteFile(filepath.Join(dir, "invalid.txt"), []byte("invalid content"), 0644)

	metas, err := store.List()
	if err != nil {
		t.Fatalf("Failed to list metas: %v", err)
	}

	if len(metas) != 2 {
		t.Errorf("Expected 2 metas, got %d", len(metas))
	}
}

// Personal.AI order the ending
