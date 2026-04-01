package cfs

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
)

type MemoryClient struct {
	mu    sync.RWMutex
	vms   map[string]*VM
	hosts map[string]*Host
}

var _ Client = &MemoryClient{}

func NewMemoryClient() *MemoryClient {
	return &MemoryClient{
		vms:   make(map[string]*VM),
		hosts: make(map[string]*Host),
	}
}

// For test injection
func (c *MemoryClient) AddVM(vm *VM) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.vms[vm.ID] = vm
}

// For test injection
func (c *MemoryClient) AddHost(host *Host) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.hosts[host.ID] = host
}

func (c *MemoryClient) ListVMs(ctx context.Context, clusterID string) ([]*VM, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []*VM
	for _, vm := range c.vms {
		if vm.ClusterID == clusterID {
			result = append(result, vm)
		}
	}
	return result, nil
}

func (c *MemoryClient) GetVM(ctx context.Context, vmID string) (*VM, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	vm, ok := c.vms[vmID]
	if !ok {
		return nil, ErrNotFound
	}
	return vm, nil
}

func (c *MemoryClient) ListHosts(ctx context.Context, clusterID string) ([]*Host, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var result []*Host
	for _, host := range c.hosts {
		if host.ClusterID == clusterID {
			result = append(result, host)
		}
	}
	return result, nil
}

func (c *MemoryClient) GetHost(ctx context.Context, hostID string) (*Host, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	host, ok := c.hosts[hostID]
	if !ok {
		return nil, ErrNotFound
	}
	return host, nil
}

// Personal.AI order the ending
