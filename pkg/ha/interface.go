package ha

import "context"

type Planner interface {
	BuildPlan(ctx context.Context, req PlanRequest) (*Plan, error)
}

type Executor interface {
	Execute(ctx context.Context, plan *Plan) error
	ExecuteWithCallback(ctx context.Context, plan *Plan, onTaskDone func(VMTask)) error
	ExecuteWithPlan(ctx context.Context, planInterface interface{}) error
}

// Personal.AI order the ending
