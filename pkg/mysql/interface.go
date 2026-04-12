package mysql

import (
	"context"
	"time"
)

type VMRepository interface {
	Upsert(ctx context.Context, record *VMRecord) error
	GetByID(ctx context.Context, vmID string) (*VMRecord, error)
	ListByCluster(ctx context.Context, clusterID string) ([]*VMRecord, error)
	ListProtected(ctx context.Context, clusterID string) ([]*VMRecord, error)
}

type HATaskRepository interface {
	Create(ctx context.Context, task *HATaskRecord) error
	UpdateStatus(ctx context.Context, taskID string, status TaskStatus) error
	ListByPlan(ctx context.Context, planID string) ([]*HATaskRecord, error)
}

type PlanRepository interface {
	Create(ctx context.Context, plan *PlanRecord) error
	GetByID(ctx context.Context, planID string) (*PlanRecord, error)
}

type Adapter interface {
	Health() MySQLStatus
	BeginTx() (TxAdapter, error)
	ClaimBoot(claim BootClaim) error
	ConfirmBoot(vmid, token string) error
	ReleaseBoot(vmid, token string) error
	GetVMState(vmid string) (*HAVMState, error)
	UpsertVMState(state HAVMState) error
	Close() error
	ListStaleBootingClaims(ctx context.Context, threshold time.Time) ([]BootClaim, error)
	ReleaseStaleClaim(ctx context.Context, vmid, token string, reason string) error
}

type TxAdapter interface {
	ClaimBoot(claim BootClaim) error
	Commit() error
	Rollback() error
}
