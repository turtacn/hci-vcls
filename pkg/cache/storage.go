package cache

type StorageStore interface {
	Get(vmid string) (*VMStorageMeta, error)
	Put(vmid string, meta VMStorageMeta) error
	Delete(vmid string) error
	List() ([]VMStorageMeta, error)
}

type StorageStoreConfig struct {
	BaseDir string
}

//Personal.AI order the ending
