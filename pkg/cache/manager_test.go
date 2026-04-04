package cache

import (
	"context"
	"errors"
	"testing"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
)

type mockStore struct {
	metas map[string]VMComputeMeta
	err   error
}

func (m *mockStore) Get(vmid string) (*VMComputeMeta, error) {
	if m.err != nil {
		return nil, m.err
	}
	meta, ok := m.metas[vmid]
	if !ok {
		return nil, ErrCacheMiss
	}
	return &meta, nil
}

func (m *mockStore) Put(vmid string, meta VMComputeMeta) error {
	m.metas[vmid] = meta
	return m.err
}

func (m *mockStore) Delete(vmid string) error {
	delete(m.metas, vmid)
	return m.err
}

func (m *mockStore) List() ([]VMComputeMeta, error) {
	var list []VMComputeMeta
	for _, meta := range m.metas {
		list = append(list, meta)
	}
	return list, m.err
}

type mockMetaSource struct {
	computeMeta *VMComputeMeta
	err         error
}

func (m *mockMetaSource) FetchVMComputeMeta(ctx context.Context, vmid string) (*VMComputeMeta, error) {
	return m.computeMeta, m.err
}

func (m *mockMetaSource) FetchVMNetworkMeta(ctx context.Context, vmid string) (*VMNetworkMeta, error) {
	return nil, m.err
}

func (m *mockMetaSource) FetchVMStorageMeta(ctx context.Context, vmid string) (*VMStorageMeta, error) {
	return nil, m.err
}

func (m *mockMetaSource) FetchVMHAMeta(ctx context.Context, vmid string) (*VMHAMeta, error) {
	return nil, m.err
}

func TestCacheManager_GetComputeMeta(t *testing.T) {
	config := CacheManagerConfig{}
	store := &mockStore{metas: make(map[string]VMComputeMeta)}
	source := &mockMetaSource{}
	log := logger.Default()
	m := metrics.NewNoopMetrics()

	mgr := NewCacheManager(config, store, nil, nil, source, log, m)
	ctx := context.Background()

	// Initial get should miss and fetch from source
	source.computeMeta = &VMComputeMeta{VMID: "100", CPUs: 2}
	meta, err := mgr.GetComputeMeta(ctx, "100")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if meta == nil || meta.CPUs != 2 {
		t.Errorf("Expected 2 CPUs, got %v", meta)
	}

	stats := mgr.Stats()
	if stats.Misses != 1 || stats.TotalEntries != 1 {
		t.Errorf("Expected 1 miss, 1 entry, got %+v", stats)
	}

	// Subsequent get should hit
	meta, err = mgr.GetComputeMeta(ctx, "100")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if meta == nil || meta.CPUs != 2 {
		t.Errorf("Expected 2 CPUs, got %v", meta)
	}

	stats = mgr.Stats()
	if stats.Hits != 1 || stats.TotalEntries != 1 {
		t.Errorf("Expected 1 hit, 1 entry, got %+v", stats)
	}
}

func TestCacheManager_StartStopSync(t *testing.T) {
	config := CacheManagerConfig{}
	mgr := NewCacheManager(config, &mockStore{}, nil, nil, &mockMetaSource{}, logger.Default(), metrics.NewNoopMetrics())
	ctx := context.Background()

	err := mgr.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed")
	}

	err = mgr.Sync(ctx, "100")
	if err != nil {
		t.Fatalf("Expected sync to succeed")
	}

	err = mgr.Stop()
	if err != nil {
		t.Fatalf("Expected stop to succeed")
	}
}

func TestCacheError(t *testing.T) {
	err := ErrCacheMiss
	if err.Error() == "" {
		t.Errorf("Expected error message to not be empty")
	}

	wrappedErr := &CacheError{Code: "ERR_CUSTOM", Message: "custom error", Err: errors.New("inner error")}
	if err := wrappedErr.Unwrap(); err == nil {
		t.Errorf("Expected to unwrap an inner error")
	}
}

// Personal.AI order the ending
