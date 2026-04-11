package app

import (
	"context"
	"github.com/turtacn/hci-vcls/pkg/cache"
	"github.com/turtacn/hci-vcls/pkg/ha"
)

// haCacheAdapter bridges cache.CacheManager (returning *VMComputeMeta)
// to ha.CacheProvider (returning interface{}).
type haCacheAdapter struct{ mgr cache.CacheManager }

// NewHACacheAdapter wraps a cache.CacheManager so it can be injected
// into ha.Executor via SetCache.
func NewHACacheAdapter(mgr cache.CacheManager) ha.CacheProvider {
	return &haCacheAdapter{mgr: mgr}
}

func (a *haCacheAdapter) GetComputeMeta(ctx context.Context, vmid string) (interface{}, error) {
	return a.mgr.GetComputeMeta(ctx, vmid)
}
