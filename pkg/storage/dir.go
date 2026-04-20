package storage

import (
	"context"
	"os"
)

type DirStorage struct {
	basePath string
}

func NewDirStorage(basePath string) *DirStorage {
	return &DirStorage{basePath: basePath}
}

func (s *DirStorage) Type() StorageType {
	return StorageDir
}

func (s *DirStorage) Probe(ctx context.Context, storageID string) (*StorageStatus, error) {
	_, err := os.Stat(s.basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &StorageStatus{
				ID:        storageID,
				Type:      StorageDir,
				Available: false,
			}, nil
		}
		return nil, err
	}

	return &StorageStatus{
		ID:        storageID,
		Type:      StorageDir,
		Available: true,
	}, nil
}

func (s *DirStorage) IsAccessible(ctx context.Context, storageID string, nodeID string) (bool, error) {
	status, err := s.Probe(ctx, storageID)
	if err != nil {
		return false, err
	}
	return status.Available, nil
}

func (s *DirStorage) Mount(ctx context.Context, storageID string, nodeID string) error {
	return nil // no-op for local dir
}

func (s *DirStorage) Unmount(ctx context.Context, storageID string, nodeID string) error {
	return nil // no-op for local dir
}
