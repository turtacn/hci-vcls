package app

import (
	"context"
	"errors"
	"testing"

	"github.com/turtacn/hci-vcls/pkg/cache"
)

type mockCacheManagerForAdapter struct {
	meta *cache.VMComputeMeta
	err  error
}

func (m *mockCacheManagerForAdapter) Start(ctx context.Context) error { return nil }
func (m *mockCacheManagerForAdapter) Stop() error                     { return nil }
func (m *mockCacheManagerForAdapter) GetComputeMeta(ctx context.Context, vmid string) (*cache.VMComputeMeta, error) {
	return m.meta, m.err
}
func (m *mockCacheManagerForAdapter) GetNetworkMeta(ctx context.Context, vmid string) (*cache.VMNetworkMeta, error) {
	return nil, nil
}
func (m *mockCacheManagerForAdapter) GetStorageMeta(ctx context.Context, vmid string) (*cache.VMStorageMeta, error) {
	return nil, nil
}
func (m *mockCacheManagerForAdapter) GetHAMeta(ctx context.Context, vmid string) (*cache.VMHAMeta, error) {
	return nil, nil
}
func (m *mockCacheManagerForAdapter) Sync(ctx context.Context, vmid string) error { return nil }
func (m *mockCacheManagerForAdapter) Stats() cache.CacheStats                     { return cache.CacheStats{} }
func (m *mockCacheManagerForAdapter) TrackVM(vmid string)                         {}
func (m *mockCacheManagerForAdapter) UntrackVM(vmid string)                       {}

func TestHACacheAdapter_GetComputeMeta(t *testing.T) {
	meta := &cache.VMComputeMeta{NodeID: "node-1"}
	mgr := &mockCacheManagerForAdapter{meta: meta, err: nil}
	adapter := NewHACacheAdapter(mgr)

	res, err := adapter.GetComputeMeta(context.Background(), "vm-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.NodeID != "node-1" {
		t.Errorf("expected node-1, got %v", res.NodeID)
	}

	mgrErr := &mockCacheManagerForAdapter{meta: nil, err: errors.New("cache miss")}
	adapterErr := NewHACacheAdapter(mgrErr)

	_, err2 := adapterErr.GetComputeMeta(context.Background(), "vm-1")
	if err2 == nil || err2.Error() != "cache miss" {
		t.Errorf("expected cache miss, got %v", err2)
	}
}
