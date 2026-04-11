package vcls

import (
	"context"
	"testing"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/cache"
	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/mysql"
)

func TestRefresh_WithCacheManager_TracksProtectedVMs(t *testing.T) {
	mockStore := NewMemoryStore()
	mockCFS := cfs.NewMemoryClient()
	mockCFS.AddVM(&cfs.VM{ID: "vm-1", ClusterID: "cluster-1", PowerState: "running"})
	mockCFS.AddVM(&cfs.VM{ID: "vm-2", ClusterID: "cluster-1", PowerState: "stopped"})

	mockRepo := mysql.NewMemoryVMRepository()
	_ = mockRepo.Upsert(context.Background(), &mysql.VMRecord{VMID: "vm-1", ClusterID: "cluster-1", Protected: true})
	log := logger.NewLogger("debug", "console")

	mockCacheManager := &mockCacheManagerImpl{
		tracked:   make(map[string]bool),
		untracked: make(map[string]bool),
	}

	service := NewServiceWithCacheManager(mockStore, mockCFS, mockRepo, nil, nil, nil, mockCacheManager, nil, log)

	err := service.Refresh(context.Background(), "cluster-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !mockCacheManager.tracked["vm-1"] {
		t.Errorf("expected vm-1 to be tracked")
	}
	if mockCacheManager.untracked["vm-1"] {
		t.Errorf("expected vm-1 to NOT be untracked")
	}

	if mockCacheManager.tracked["vm-2"] {
		t.Errorf("expected vm-2 to NOT be tracked")
	}
	if !mockCacheManager.untracked["vm-2"] {
		t.Errorf("expected vm-2 to be untracked")
	}
}

func TestRefresh_WithoutCacheManager_NoOp(t *testing.T) {
	mockStore := NewMemoryStore()
	mockCFS := cfs.NewMemoryClient()
	mockCFS.AddVM(&cfs.VM{ID: "vm-1", ClusterID: "cluster-1", PowerState: "running"})

	mockRepo := mysql.NewMemoryVMRepository()
	_ = mockRepo.Upsert(context.Background(), &mysql.VMRecord{VMID: "vm-1", ClusterID: "cluster-1", Protected: true})
	log := logger.NewLogger("debug", "console")

	service := NewService(mockStore, mockCFS, mockRepo, nil, nil, nil, nil, log)

	err := service.Refresh(context.Background(), "cluster-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// should not panic
}

type mockCacheManagerImpl struct {
	tracked   map[string]bool
	untracked map[string]bool
}

func (m *mockCacheManagerImpl) Start(ctx context.Context) error { return nil }
func (m *mockCacheManagerImpl) Stop() error                     { return nil }
func (m *mockCacheManagerImpl) GetComputeMeta(ctx context.Context, vmid string) (*cache.VMComputeMeta, error) {
	return nil, nil
}
func (m *mockCacheManagerImpl) GetNetworkMeta(ctx context.Context, vmid string) (*cache.VMNetworkMeta, error) {
	return nil, nil
}
func (m *mockCacheManagerImpl) GetStorageMeta(ctx context.Context, vmid string) (*cache.VMStorageMeta, error) {
	return nil, nil
}
func (m *mockCacheManagerImpl) GetHAMeta(ctx context.Context, vmid string) (*cache.VMHAMeta, error) {
	return nil, nil
}
func (m *mockCacheManagerImpl) Sync(ctx context.Context, vmid string) error { return nil }
func (m *mockCacheManagerImpl) Stats() cache.CacheStats                     { return cache.CacheStats{} }
func (m *mockCacheManagerImpl) TrackVM(vmid string)                         { m.tracked[vmid] = true }
func (m *mockCacheManagerImpl) UntrackVM(vmid string)                       { m.untracked[vmid] = true }
