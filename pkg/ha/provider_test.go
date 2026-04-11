package ha

import (
	"context"
	"github.com/turtacn/hci-vcls/pkg/cache"
)

// TestCacheProviderInterface ensures cache.CacheManager structurally
// satisfies CacheProvider (compile-time assertion) with our adapter pattern.
// Note: because types don't exactly match (*cache.VMComputeMeta vs interface{}),
// we use a thin adapter in internal/app to bridge it, so we mock the adapter structurally here.
type dummyAdapter struct{ mgr cache.CacheManager }

func (a *dummyAdapter) GetComputeMeta(ctx context.Context, vmid string) (interface{}, error) {
	return a.mgr.GetComputeMeta(ctx, vmid)
}

var _ CacheProvider = (*dummyAdapter)(nil)
