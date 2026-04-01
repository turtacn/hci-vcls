package cache

import (
	"context"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
)

type cacheManagerImpl struct {
	config     CacheManagerConfig
	cStore     Store
	nStore     NetworkStore
	sStore     StorageStore
	metaSource MetaSource
	log        logger.Logger
	metrics    metrics.Metrics
	stats      CacheStats
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewCacheManager(config CacheManagerConfig, c Store, n NetworkStore, s StorageStore, source MetaSource, log logger.Logger, m metrics.Metrics) CacheManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &cacheManagerImpl{
		config:     config,
		cStore:     c,
		nStore:     n,
		sStore:     s,
		metaSource: source,
		log:        log,
		metrics:    m,
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (m *cacheManagerImpl) Start(ctx context.Context) error {
	// Start sync loop
	return nil
}

func (m *cacheManagerImpl) Stop() error {
	m.cancel()
	return nil
}

func (m *cacheManagerImpl) GetComputeMeta(ctx context.Context, vmid string) (*VMComputeMeta, error) {
	meta, err := m.cStore.Get(vmid)
	if err == ErrCacheMiss && m.metaSource != nil {
		m.stats.Misses++
		meta, err = m.metaSource.FetchVMComputeMeta(ctx, vmid)
		if err == nil && meta != nil {
			m.cStore.Put(vmid, *meta)
			m.stats.TotalEntries++
		}
	} else if err == nil {
		m.stats.Hits++
	}
	return meta, err
}

func (m *cacheManagerImpl) GetNetworkMeta(ctx context.Context, vmid string) (*VMNetworkMeta, error) {
	if m.nStore == nil {
		return nil, nil
	}
	meta, err := m.nStore.Get(vmid)
	if err == ErrCacheMiss && m.metaSource != nil {
		m.stats.Misses++
		meta, err = m.metaSource.FetchVMNetworkMeta(ctx, vmid)
		if err == nil && meta != nil {
			m.nStore.Put(vmid, *meta)
			m.stats.TotalEntries++
		}
	} else if err == nil {
		m.stats.Hits++
	}
	return meta, err
}

func (m *cacheManagerImpl) GetStorageMeta(ctx context.Context, vmid string) (*VMStorageMeta, error) {
	if m.sStore == nil {
		return nil, nil
	}
	meta, err := m.sStore.Get(vmid)
	if err == ErrCacheMiss && m.metaSource != nil {
		m.stats.Misses++
		meta, err = m.metaSource.FetchVMStorageMeta(ctx, vmid)
		if err == nil && meta != nil {
			m.sStore.Put(vmid, *meta)
			m.stats.TotalEntries++
		}
	} else if err == nil {
		m.stats.Hits++
	}
	return meta, err
}

func (m *cacheManagerImpl) GetHAMeta(ctx context.Context, vmid string) (*VMHAMeta, error) {
	// MVP uses only compute/network/storage metas, ha meta is for future expansion
	return nil, nil
}

func (m *cacheManagerImpl) Sync(ctx context.Context, vmid string) error {
	// For MVP, explicitly forcing cache refresh
	if m.metaSource == nil {
		return nil
	}
	cMeta, _ := m.metaSource.FetchVMComputeMeta(ctx, vmid)
	if cMeta != nil && m.cStore != nil {
		m.cStore.Put(vmid, *cMeta)
	}

	nMeta, _ := m.metaSource.FetchVMNetworkMeta(ctx, vmid)
	if nMeta != nil && m.nStore != nil {
		m.nStore.Put(vmid, *nMeta)
	}

	sMeta, _ := m.metaSource.FetchVMStorageMeta(ctx, vmid)
	if sMeta != nil && m.sStore != nil {
		m.sStore.Put(vmid, *sMeta)
	}
	return nil
}

func (m *cacheManagerImpl) Stats() CacheStats {
	return m.stats
}

//Personal.AI order the ending
