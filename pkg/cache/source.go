package cache

import "context"

type MetaSource interface {
	FetchVMComputeMeta(ctx context.Context, vmid string) (*VMComputeMeta, error)
	FetchVMNetworkMeta(ctx context.Context, vmid string) (*VMNetworkMeta, error)
	FetchVMStorageMeta(ctx context.Context, vmid string) (*VMStorageMeta, error)
	FetchVMHAMeta(ctx context.Context, vmid string) (*VMHAMeta, error)
}

type MultiSource struct {
	Primary MetaSource
	Backup  MetaSource
}

func (m *MultiSource) FetchVMComputeMeta(ctx context.Context, vmid string) (*VMComputeMeta, error) {
	meta, err := m.Primary.FetchVMComputeMeta(ctx, vmid)
	if err == nil {
		return meta, nil
	}
	if m.Backup != nil {
		return m.Backup.FetchVMComputeMeta(ctx, vmid)
	}
	return nil, err
}

func (m *MultiSource) FetchVMNetworkMeta(ctx context.Context, vmid string) (*VMNetworkMeta, error) {
	meta, err := m.Primary.FetchVMNetworkMeta(ctx, vmid)
	if err == nil {
		return meta, nil
	}
	if m.Backup != nil {
		return m.Backup.FetchVMNetworkMeta(ctx, vmid)
	}
	return nil, err
}

func (m *MultiSource) FetchVMStorageMeta(ctx context.Context, vmid string) (*VMStorageMeta, error) {
	meta, err := m.Primary.FetchVMStorageMeta(ctx, vmid)
	if err == nil {
		return meta, nil
	}
	if m.Backup != nil {
		return m.Backup.FetchVMStorageMeta(ctx, vmid)
	}
	return nil, err
}

func (m *MultiSource) FetchVMHAMeta(ctx context.Context, vmid string) (*VMHAMeta, error) {
	meta, err := m.Primary.FetchVMHAMeta(ctx, vmid)
	if err == nil {
		return meta, nil
	}
	if m.Backup != nil {
		return m.Backup.FetchVMHAMeta(ctx, vmid)
	}
	return nil, err
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

