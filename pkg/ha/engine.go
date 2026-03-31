package ha

import "context"

type HAEngine interface {
	Start(ctx context.Context) error
	Stop() error
	Evaluate(ctx context.Context, vmid string) (HADecision, error)
	Execute(ctx context.Context, decision HADecision) error
	BatchBoot(ctx context.Context, decisions []HADecision, policy BatchBootPolicy) error
	ClusterView() ClusterView
	ActiveTasks() map[string]BootTask
	OnDecision(callback func(decision HADecision))
}

//Personal.AI order the ending