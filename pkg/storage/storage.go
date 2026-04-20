package storage

import "context"

type StorageType string

const (
	StorageRBD   StorageType = "rbd"
	StorageNFS   StorageType = "nfs"
	StorageISCSI StorageType = "iscsi"
	StorageDir   StorageType = "dir"
	StorageZFS   StorageType = "zfs"
)

type StorageStatus struct {
	ID        string
	Type      StorageType
	Available bool
	UsedBytes int64
	FreeBytes int64
	NodeID    string
}

type ExternalStorage interface {
	Type() StorageType
	Probe(ctx context.Context, storageID string) (*StorageStatus, error)
	IsAccessible(ctx context.Context, storageID string, nodeID string) (bool, error)
	Mount(ctx context.Context, storageID string, nodeID string) error
	Unmount(ctx context.Context, storageID string, nodeID string) error
}
