package cache

import (
	"testing"
)

func TestStorageStore_CRUD(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewStorageStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create storage store: %v", err)
	}

	meta := VMStorageMeta{
		VMID: "100",
		Disks: []DiskConfig{
			{DiskID: "disk-1", SizeGB: 10},
		},
	}

	err = store.Put("100", meta)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	retrieved, err := store.Get("100")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(retrieved.Disks) != 1 {
		t.Errorf("Expected 1 disk, got %d", len(retrieved.Disks))
	}

	list, err := store.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("Expected list length 1, got %d", len(list))
	}

	err = store.Delete("100")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = store.Get("100")
	if err != ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss after delete, got %v", err)
	}
}
