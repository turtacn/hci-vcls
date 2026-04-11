package cache

import "context"

type CacheStats struct {
	TotalEntries  int
	Hits          int
	Misses        int
	Evictions     int
	SyncFailures  int
	SyncSuccesses int
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

	// TrackVM marks a VM ID as a sync target for the background sync loop.
	// Safe to call concurrently.
	TrackVM(vmid string)

	// UntrackVM removes a VM ID from the sync target set.
	// Safe to call concurrently. No-op if the VM was not tracked.
	UntrackVM(vmid string)
}

