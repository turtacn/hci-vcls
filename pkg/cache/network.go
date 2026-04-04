package cache

type NetworkStore interface {
	Get(vmid string) (*VMNetworkMeta, error)
	Put(vmid string, meta VMNetworkMeta) error
	Delete(vmid string) error
	List() ([]VMNetworkMeta, error)
}

type NetworkStoreConfig struct {
	BaseDir string
}

// Personal.AI order the ending
