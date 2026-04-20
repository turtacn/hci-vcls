package ha

import (
	"context"

	"github.com/turtacn/hci-vcls/pkg/cache"
)

// CacheProvider supplies VM metadata from a local cache for minority
// boot paths. Implementations typically live in pkg/cache.
type CacheProvider interface {
	GetComputeMeta(ctx context.Context, vmid string) (*cache.VMComputeMeta, error)
}

// StateProvider supplies current cluster degradation level for the
// executor's execution gate.
type StateProvider interface {
	CurrentLevel() string
}

type StorageReachabilityProbe interface {
	ProbeAll(ctx context.Context, storageID string) ([]*StorageStatus, error)
}

type StorageStatus struct {
	Available bool
}

type Planner interface {
	BuildPlan(ctx context.Context, req PlanRequest) (*Plan, error)
}

// ExecuteOpts configures a single Execute invocation.
// Zero value means normal production execution.
type ExecuteOpts struct {
	DryRun         bool
	MaxConcurrency int
	Timeout        int
}

type Executor interface {
	Execute(ctx context.Context, plan *Plan, opts ExecuteOpts) error
	ExecuteWithCallback(ctx context.Context, plan *Plan, opts ExecuteOpts, onTaskDone func(VMTask)) error
	ExecuteWithPlan(ctx context.Context, plan *Plan, opts ExecuteOpts) error
}

type AuditSink interface {
	LogHADecision(ctx context.Context, clusterID, vmid, planID, bootPath, sourceHost, targetHost, reason, degradation, outcome, errStr string, dryRun bool) error
}
