package cfs

import "context"

type Client interface {
	ListVMs(ctx context.Context, clusterID string) ([]*VM, error)
	GetVM(ctx context.Context, vmID string) (*VM, error)
	ListHosts(ctx context.Context, clusterID string) ([]*Host, error)
	GetHost(ctx context.Context, hostID string) (*Host, error)
}

// Personal.AI order the ending
