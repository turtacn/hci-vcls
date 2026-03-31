package cache

import "context"

type MetaSource interface {
	FetchVMComputeMeta(ctx context.Context, vmid string) (*VMComputeMeta, error)
	FetchVMNetworkMeta(ctx context.Context, vmid string) (*VMNetworkMeta, error)
	FetchVMStorageMeta(ctx context.Context, vmid string) (*VMStorageMeta, error)
	FetchVMHAMeta(ctx context.Context, vmid string) (*VMHAMeta, error)
}

type CFSMetaSource struct {
	// dependencies on pkg/cfs, logger
}

type MySQLMetaSource struct {
	// dependencies on pkg/mysql, logger
}

func ParsePVEConfig(raw []byte) (VMHAMeta, error) {
	// minimal parse logic for VMHAMeta
	return VMHAMeta{}, nil
}

//Personal.AI order the ending