package qm

import "context"

type Executor interface {
	Start(ctx context.Context, vmid string, opts BootOptions) BootResult
	Status(ctx context.Context, vmid string) (VMStatus, error)
	Stop(ctx context.Context, vmid string, opts BootOptions) error
	Lock(ctx context.Context, vmid string) error
	Unlock(ctx context.Context, vmid string) error
}

//Personal.AI order the ending
