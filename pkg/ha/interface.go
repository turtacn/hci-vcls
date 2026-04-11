package ha

import "context"

// CacheProvider supplies VM metadata from a local cache for minority
// boot paths. Implementations typically live in pkg/cache.
type CacheProvider interface {
	GetComputeMeta(ctx context.Context, vmid string) (interface{}, error)
}

// StateProvider supplies current cluster degradation level for the
// executor's execution gate.
type StateProvider interface {
	CurrentLevel() string
}

type Planner interface {
	BuildPlan(ctx context.Context, req PlanRequest) (*Plan, error)
}

type Executor interface {
	Execute(ctx context.Context, plan *Plan) error
	ExecuteWithCallback(ctx context.Context, plan *Plan, onTaskDone func(VMTask)) error
	ExecuteWithPlan(ctx context.Context, planInterface interface{}) error
}

