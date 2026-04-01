package cfs

import (
	"context"
	"testing"
)

func TestMemoryClient_VMs(t *testing.T) {
	c := NewMemoryClient()
	ctx := context.Background()

	vm := &VM{ID: "vm1", ClusterID: "c1", PowerState: "running"}
	c.AddVM(vm)

	found, err := c.GetVM(ctx, "vm1")
	if err != nil || found.PowerState != "running" {
		t.Errorf("GetVM failed: %v", err)
	}

	list, err := c.ListVMs(ctx, "c1")
	if err != nil || len(list) != 1 {
		t.Errorf("ListVMs failed: %v", err)
	}

	_, err = c.GetVM(ctx, "nonexistent")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound")
	}
}

func TestMemoryClient_Hosts(t *testing.T) {
	c := NewMemoryClient()
	ctx := context.Background()

	host := &Host{ID: "h1", ClusterID: "c1", Healthy: true}
	c.AddHost(host)

	found, err := c.GetHost(ctx, "h1")
	if err != nil || !found.Healthy {
		t.Errorf("GetHost failed: %v", err)
	}

	list, err := c.ListHosts(ctx, "c1")
	if err != nil || len(list) != 1 {
		t.Errorf("ListHosts failed: %v", err)
	}
}

// Personal.AI order the ending
