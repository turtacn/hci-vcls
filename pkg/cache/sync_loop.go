package cache

import (
	"time"
)

// syncLoop starts a background goroutine to periodically refresh all protected VMs' metadata.
func (m *cacheManagerImpl) syncLoop() {
	if m.config.SyncIntervalMs <= 0 {
		m.config.SyncIntervalMs = 30000 // default 30s
	}

	ticker := time.NewTicker(time.Duration(m.config.SyncIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	// Optionally do an initial sync right away.
	m.runSyncAll()

	for {
		select {
		case <-m.ctx.Done():
			if m.log != nil {
				m.log.Info("CacheManager sync loop stopping")
			}
			return
		case <-ticker.C:
			m.runSyncAll()
		}
	}
}

// runSyncAll is a helper to fetch list of VMs and sync each.
func (m *cacheManagerImpl) runSyncAll() {
	// In a complete system, CacheManager might receive a list of VMs to track,
	// or ask VCLS service. For now, we will sync all keys currently in the compute cache
	// or rely on an injected VM list provider.
	// As a fallback, we fetch list of cached VMs and refresh them.
	vmsToSync := m.getVMsToSync()

	for _, vmid := range vmsToSync {
		// Single failure should not stop the loop
		err := m.Sync(m.ctx, vmid)

		m.mu.Lock()
		if err != nil {
			m.stats.SyncFailures++
			if m.log != nil {
				m.log.Warn("Failed to sync VM cache", "vm", vmid, "error", err)
			}
		} else {
			m.stats.SyncSuccesses++
		}
		m.mu.Unlock()
	}
}

func (m *cacheManagerImpl) getVMsToSync() []string {
	// For demonstration, retrieve all keys from compute store
	// (Assuming the store interface can list or provide keys; if not we might need to track them).
	// If the current interface doesn't support Listing keys, we add a TODO.
	// TODO: Replace with an actual list of protected VMs provided by vCLS.

	// Right now we don't have a direct ListKeys, so we'll just return an empty slice unless there's a mechanism.
	// Actually, `mockStore` in manager_test.go has `List()` which implies Store has it or can have it.
	// Let's assume we maintain a list of tracked VMs in the manager if needed, or query store.

	m.mu.RLock()
	defer m.mu.RUnlock()

	vms := make([]string, 0, len(m.trackedVMs))
	for vmid := range m.trackedVMs {
		vms = append(vms, vmid)
	}

	return vms
}

// TrackVM adds a VM to the sync loop tracking.
func (m *cacheManagerImpl) TrackVM(vmid string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.trackedVMs == nil {
		m.trackedVMs = make(map[string]bool)
	}
	m.trackedVMs[vmid] = true
}

// UntrackVM removes a VM from the sync loop tracking.
func (m *cacheManagerImpl) UntrackVM(vmid string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.trackedVMs != nil {
		delete(m.trackedVMs, vmid)
	}
}
