package cache

import (
	"testing"
)

func TestNetworkStore_CRUD(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewNetworkStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to create network store: %v", err)
	}

	meta := VMNetworkMeta{
		VMID: "100",
		NICs: []NICConfig{
			{MACAddress: "00:11:22"},
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
	if len(retrieved.NICs) != 1 {
		t.Errorf("Expected 1 NIC, got %d", len(retrieved.NICs))
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
