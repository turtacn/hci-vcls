package storage

import (
	"context"
	"fmt"
	"sync"
)

type Registry interface {
	Register(storage ExternalStorage)
	Get(t StorageType) (ExternalStorage, error)
	ProbeAll(ctx context.Context, storageID string) ([]*StorageStatus, error)
}

type registryImpl struct {
	mu       sync.RWMutex
	backends map[StorageType]ExternalStorage
}

func NewRegistry() Registry {
	return &registryImpl{
		backends: make(map[StorageType]ExternalStorage),
	}
}

func (r *registryImpl) Register(storage ExternalStorage) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.backends[storage.Type()] = storage
}

func (r *registryImpl) Get(t StorageType) (ExternalStorage, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if s, ok := r.backends[t]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("storage type %s not registered", t)
}

func (r *registryImpl) ProbeAll(ctx context.Context, storageID string) ([]*StorageStatus, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var statuses []*StorageStatus
	for _, s := range r.backends {
		status, err := s.Probe(ctx, storageID)
		if err == nil {
			statuses = append(statuses, status)
		}
	}
	return statuses, nil
}
