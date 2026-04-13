package mysql

import (
	"context"
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
