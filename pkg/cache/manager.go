package cache

import "context"

type CacheStats struct {
	TotalEntries int
	Hits         int
	Misses       int
	Evictions    int
}

type CacheManagerConfig struct {
	BaseDir           string
	MaxConcurrentSync int
	SyncIntervalMs    int
	TTLMs             int
}

type CacheManager interface {
	Start(ctx context.Context) error
	Stop() error
	GetComputeMeta(ctx context.Context, vmid string) (*VMComputeMeta, error)
	GetNetworkMeta(ctx context.Context, vmid string) (*VMNetworkMeta, error)
	GetStorageMeta(ctx context.Context, vmid string) (*VMStorageMeta, error)
	GetHAMeta(ctx context.Context, vmid string) (*VMHAMeta, error)
	Sync(ctx context.Context, vmid string) error
	Stats() CacheStats
}

//Personal.AI order the ending
