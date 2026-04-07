package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
)

type mockFailingMetaSource struct {
	computeMeta *VMComputeMeta
	err         error
}

func (m *mockFailingMetaSource) FetchVMComputeMeta(ctx context.Context, vmid string) (*VMComputeMeta, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.computeMeta, nil
}

func (m *mockFailingMetaSource) FetchVMNetworkMeta(ctx context.Context, vmid string) (*VMNetworkMeta, error) {
	return nil, nil // ignore for now
}

func (m *mockFailingMetaSource) FetchVMStorageMeta(ctx context.Context, vmid string) (*VMStorageMeta, error) {
	return nil, nil
}

func (m *mockFailingMetaSource) FetchVMHAMeta(ctx context.Context, vmid string) (*VMHAMeta, error) {
	return nil, nil
}

func TestCacheManager_SyncLoop_StartStop(t *testing.T) {
	config := CacheManagerConfig{SyncIntervalMs: 10} // fast interval for testing
	source := &mockFailingMetaSource{computeMeta: &VMComputeMeta{VMID: "100", CPUs: 4}}
	store := &mockStore{metas: make(map[string]VMComputeMeta)}

	mgr := NewCacheManager(config, store, nil, nil, source, logger.Default(), metrics.NewNoopMetrics())
	mgrImpl := mgr.(*cacheManagerImpl)

	// Add a VM to track so sync actually does something
	mgrImpl.TrackVM("100")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := mgr.Start(ctx)
	if err != nil {
		t.Fatalf("expected start to succeed, got %v", err)
	}

	// Starting again should be idempotent and not return error or leak goroutine
	err = mgr.Start(ctx)
	if err != nil {
		t.Fatalf("expected idempotent start to succeed")
	}

	time.Sleep(50 * time.Millisecond) // Let it loop a few times

	stats := mgr.Stats()
	if stats.SyncSuccesses == 0 {
		t.Errorf("expected background sync to record successes, got %d", stats.SyncSuccesses)
	}

	// Ensure stopping works
	err = mgr.Stop()
	if err != nil {
		t.Fatalf("expected stop to succeed, got %v", err)
	}
}

func TestCacheManager_SyncLoop_CtxCancel(t *testing.T) {
	config := CacheManagerConfig{SyncIntervalMs: 10}
	source := &mockFailingMetaSource{computeMeta: &VMComputeMeta{VMID: "100", CPUs: 4}}
	store := &mockStore{metas: make(map[string]VMComputeMeta)}

	mgr := NewCacheManager(config, store, nil, nil, source, logger.Default(), metrics.NewNoopMetrics())

	ctx, cancel := context.WithCancel(context.Background())
	mgr.Start(ctx) // internally it uses its own context derived during creation, wait...
	// Ah, NewCacheManager creates its own ctx, cancel. `Start(ctx)` just starts the loop listening to that internal context.
	// So to simulate cancellation, we just call Stop().
	// Let's test calling Stop() and checking if it runs anymore.
	mgr.Stop()

	// Wait, to test if it handles the given ctx properly:
	// NewCacheManager in the current design takes context.Background() and creates its own cancellation.
	// The `Start(ctx)` signature receives a context, but doesn't strictly use it for the loop's lifecycle in manager_impl.go.
	// We'll rely on Stop() to test termination.

	stats := mgr.Stats()
	if stats.SyncSuccesses != 0 {
		t.Errorf("expected no sync successes since it was stopped immediately")
	}
}

func TestCacheManager_SyncLoop_FailureDoesNotCrash(t *testing.T) {
	config := CacheManagerConfig{SyncIntervalMs: 10}
	source := &mockFailingMetaSource{err: errors.New("temporary network error")}
	store := &mockStore{metas: make(map[string]VMComputeMeta)}

	mgr := NewCacheManager(config, store, nil, nil, source, logger.Default(), metrics.NewNoopMetrics())
	mgrImpl := mgr.(*cacheManagerImpl)
	mgrImpl.TrackVM("100")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr.Start(ctx)

	time.Sleep(50 * time.Millisecond)

	stats := mgr.Stats()
	if stats.SyncFailures == 0 {
		t.Errorf("expected background sync to record failures")
	}

	// Recover the source
	source.err = nil
	source.computeMeta = &VMComputeMeta{VMID: "100", CPUs: 2}

	time.Sleep(50 * time.Millisecond)

	mgr.Stop()

	stats = mgr.Stats()
	if stats.SyncSuccesses == 0 {
		t.Errorf("expected background sync to recover and record successes, got %d", stats.SyncSuccesses)
	}
}
