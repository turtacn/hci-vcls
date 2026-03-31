package ha

import "context"

type Evaluator interface {
	Evaluate(ctx context.Context, vmid string, cv ClusterView) (HADecision, error)
	SelectBootPath(ctx context.Context, vmid string) (BootPath, error)
	SelectTargetNode(ctx context.Context, vmid string, cv ClusterView) (string, error)
	AssignPriority(ctx context.Context, vmid string) (int, error)
}

//Personal.AI order the ending