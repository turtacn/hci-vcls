package heartbeat

import (
	"sync"
	"time"
)

type memoryMonitor struct {
	mu           sync.RWMutex
	summaries    map[string]Summary
	storageLayer *StorageHeartbeater
}

var _ Monitor = &memoryMonitor{}

func NewMemoryMonitor(storageLayer *StorageHeartbeater) *memoryMonitor {
	return &memoryMonitor{
		summaries:    make(map[string]Summary),
		storageLayer: storageLayer,
	}
}

func (m *memoryMonitor) Record(sample Sample) {
	m.mu.Lock()
	defer m.mu.Unlock()

	summary, exists := m.summaries[sample.NodeID]
	if !exists {
		summary = Summary{
			NodeID:    sample.NodeID,
			ClusterID: sample.ClusterID,
		}
	}

	summary.LastSeenAt = sample.ReceivedAt
	summary.Healthy = true
	summary.ObservedAt = time.Now()
	m.summaries[sample.NodeID] = summary
}

func (m *memoryMonitor) GetSummary(nodeID string) (Summary, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, exists := m.summaries[nodeID]
	return s, exists
}

func (m *memoryMonitor) ListSummaries(clusterID string) []Summary {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []Summary
	for _, s := range m.summaries {
		if s.ClusterID == clusterID {
			result = append(result, s)
		}
	}
	return result
}

func (m *memoryMonitor) CheckTimeouts(now time.Time, timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for nodeID, s := range m.summaries {
		if s.Healthy && now.Sub(s.LastSeenAt) > timeout {
			// Complement L1 UDP loss with L2 Storage Heartbeat if available
			if m.storageLayer != nil {
				if ts, err := m.storageLayer.Read(nodeID); err == nil {
					// Check if storage heartbeat is fresh
					if now.Sub(ts) <= timeout {
						s.LastSeenAt = ts
						m.summaries[nodeID] = s
						continue
					}
				}
			}

			s.Healthy = false
			s.LostCount++
			m.summaries[nodeID] = s
		}
	}
}

