package vcls

import (
	"sync"
	"time"
)

type memoryStore struct {
	mu  sync.RWMutex
	vms map[string]*VM
}

var _ Store = &memoryStore{}

func NewMemoryStore() *memoryStore {
	return &memoryStore{
		vms: make(map[string]*VM),
	}
}

func (s *memoryStore) Put(vm *VM) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.vms[vm.ID] = vm
}

func (s *memoryStore) Get(vmID string) (*VM, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	vm, ok := s.vms[vmID]
	if !ok {
		return nil, false
	}
	vmCopy := *vm
	return &vmCopy, true
}

func (s *memoryStore) List(clusterID string) []*VM {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*VM
	for _, vm := range s.vms {
		if vm.ClusterID == clusterID {
			vmCopy := *vm
			result = append(result, &vmCopy)
		}
	}
	return result
}

func (s *memoryStore) ListEligible(clusterID string) []*VM {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []*VM
	for _, vm := range s.vms {
		if vm.ClusterID == clusterID && vm.EligibleForHA {
			vmCopy := *vm
			result = append(result, &vmCopy)
		}
	}
	return result
}

func (s *memoryStore) Status(clusterID string) Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var vmCount, protectedCount, eligibleCount int
	var lastRefresh time.Time

	for _, vm := range s.vms {
		if vm.ClusterID != clusterID {
			continue
		}
		vmCount++
		if vm.Protected {
			protectedCount++
		}
		if vm.EligibleForHA {
			eligibleCount++
		}
		if vm.LastSyncAt.After(lastRefresh) {
			lastRefresh = vm.LastSyncAt
		}
	}

	return Status{
		ClusterID:      clusterID,
		VMCount:        vmCount,
		ProtectedCount: protectedCount,
		EligibleCount:  eligibleCount,
		LastRefreshAt:  lastRefresh,
	}
}

