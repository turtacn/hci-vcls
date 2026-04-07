package zk

import (
	"context"
	"testing"
	"time"
)

func TestMemoryClient_NotConnected(t *testing.T) {
	client := NewMemoryClient()

	_, err := client.Get("/test")
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got %v", err)
	}

	err = client.Create("/test", "data", false)
	if err != ErrNotConnected {
		t.Errorf("Expected ErrNotConnected, got %v", err)
	}
}

func TestMemoryClient_CRUD(t *testing.T) {
	client := NewMemoryClient()
	_ = client.Connect(context.Background())

	err := client.Create("/node1", "val1", false)
	if err != nil {
		t.Fatalf("Failed to create node: %v", err)
	}

	val, err := client.Get("/node1")
	if err != nil || val != "val1" {
		t.Errorf("Get failed or returned wrong value: %v, val: %s", err, val)
	}

	err = client.Set("/node1", "val2", 0)
	if err != nil {
		t.Errorf("Set failed: %v", err)
	}

	val, _ = client.Get("/node1")
	if val != "val2" {
		t.Errorf("Set did not update value, got %s", val)
	}

	err = client.Delete("/node1", 1)
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	exists, _ := client.Exists("/node1")
	if exists {
		t.Errorf("Node should be deleted")
	}
}

func TestMemoryClient_Watch(t *testing.T) {
	client := NewMemoryClient()
	_ = client.Connect(context.Background())

	_ = client.Create("/watch-node", "v1", false)

	watchCh, err := client.Watch("/watch-node")
	if err != nil {
		t.Fatalf("Failed to watch: %v", err)
	}

	_ = client.Set("/watch-node", "v2", 0)

	select {
	case event := <-watchCh:
		if event.Path != "/watch-node" || event.Type != EventNodeDataChanged {
			t.Errorf("Unexpected event: %+v", event)
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Did not receive watch event")
	}
}

func TestMemoryClient_CreateExists(t *testing.T) {
	client := NewMemoryClient()
	_ = client.Connect(context.Background())

	_ = client.Create("/exist-node", "v1", false)
	err := client.Create("/exist-node", "v2", false)
	if err != ErrNodeExists {
		t.Errorf("Expected ErrNodeExists, got %v", err)
	}
}

