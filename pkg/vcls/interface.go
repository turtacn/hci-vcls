package vcls

import "context"

type Store interface {
	Put(vm *VM)
	Get(vmID string) (*VM, bool)
	List(clusterID string) []*VM
	ListEligible(clusterID string) []*VM
	Status(clusterID string) Status
}

type Service interface {
	Refresh(ctx context.Context, clusterID string) error
	GetVM(ctx context.Context, vmID string) (*VM, error)
	ListProtected(ctx context.Context, clusterID string) ([]*VM, error)
	ListEligible(ctx context.Context, clusterID string) ([]*VM, error)
}

