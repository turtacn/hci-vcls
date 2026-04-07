package cache

import (
	"context"
	"sync"
	"sync/atomic"

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
	mu         sync.RWMutex
	trackedVMs map[string]bool
	ctx        context.Context
	cancel     context.CancelFunc
	started    atomic.Bool
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
		trackedVMs: make(map[string]bool),
		ctx:        ctx,
		cancel:     cancel,
	}
}

func (m *cacheManagerImpl) Start(ctx context.Context) error {
	if m.started.CompareAndSwap(false, true) {
		go m.syncLoop()
	}
	return nil
}

func (m *cacheManagerImpl) Stop() error {
	if m.started.CompareAndSwap(true, false) {
		m.cancel()
	}
	return nil
}

func (m *cacheManagerImpl) GetComputeMeta(ctx context.Context, vmid string) (*VMComputeMeta, error) {
	meta, err := m.cStore.Get(vmid)
	m.mu.Lock()
	if err == ErrCacheMiss && m.metaSource != nil {
		m.stats.Misses++
		m.mu.Unlock()
		meta, err = m.metaSource.FetchVMComputeMeta(ctx, vmid)
		if err == nil && meta != nil {
			_ = m.cStore.Put(vmid, *meta)
			m.mu.Lock()
			m.stats.TotalEntries++
			m.mu.Unlock()
		}
	} else if err == nil {
		m.stats.Hits++
		m.mu.Unlock()
	} else {
		m.mu.Unlock()
	}
	return meta, err
}

func (m *cacheManagerImpl) GetNetworkMeta(ctx context.Context, vmid string) (*VMNetworkMeta, error) {
	meta, err := m.nStore.Get(vmid)
	m.mu.Lock()
	if err == ErrCacheMiss && m.metaSource != nil {
		m.stats.Misses++
		m.mu.Unlock()
		meta, err = m.metaSource.FetchVMNetworkMeta(ctx, vmid)
		if err == nil && meta != nil {
			_ = m.nStore.Put(vmid, *meta)
			m.mu.Lock()
			m.stats.TotalEntries++
			m.mu.Unlock()
		}
	} else if err == nil {
		m.stats.Hits++
		m.mu.Unlock()
	} else {
		m.mu.Unlock()
	}
	return meta, err
}

func (m *cacheManagerImpl) GetStorageMeta(ctx context.Context, vmid string) (*VMStorageMeta, error) {
	meta, err := m.sStore.Get(vmid)
	m.mu.Lock()
	if err == ErrCacheMiss && m.metaSource != nil {
		m.stats.Misses++
		m.mu.Unlock()
		meta, err = m.metaSource.FetchVMStorageMeta(ctx, vmid)
		if err == nil && meta != nil {
			_ = m.sStore.Put(vmid, *meta)
			m.mu.Lock()
			m.stats.TotalEntries++
			m.mu.Unlock()
		}
	} else if err == nil {
		m.stats.Hits++
		m.mu.Unlock()
	} else {
		m.mu.Unlock()
	}
	return meta, err
}

func (m *cacheManagerImpl) GetHAMeta(ctx context.Context, vmid string) (*VMHAMeta, error) {
	if m.metaSource != nil {
		return m.metaSource.FetchVMHAMeta(ctx, vmid)
	}
	return nil, ErrCacheMiss
}

func (m *cacheManagerImpl) Sync(ctx context.Context, vmid string) error {
	if m.metaSource == nil {
		return nil
	}

	// Fetch Compute Meta
	cMeta, err := m.metaSource.FetchVMComputeMeta(ctx, vmid)
	if err == nil && cMeta != nil {
		m.cStore.Put(vmid, *cMeta)
	}

	// Fetch Network Meta
	nMeta, err := m.metaSource.FetchVMNetworkMeta(ctx, vmid)
	if err == nil && nMeta != nil {
		m.nStore.Put(vmid, *nMeta)
	}

	// Fetch Storage Meta
	sMeta, err := m.metaSource.FetchVMStorageMeta(ctx, vmid)
	if err == nil && sMeta != nil {
		m.sStore.Put(vmid, *sMeta)
	}

	return nil
}

func (m *cacheManagerImpl) Stats() CacheStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

