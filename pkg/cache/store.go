package cache

type Store interface {
	Get(vmid string) (*VMComputeMeta, error)
	Put(vmid string, meta VMComputeMeta) error
	Delete(vmid string) error
	List() ([]VMComputeMeta, error)
}

//Personal.AI order the ending