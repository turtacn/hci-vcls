package fdm

import "context"

type Agent interface {
	Start(ctx context.Context) error
	Stop() error
	NodeStates() map[string]NodeState
	LocalDegradationLevel() DegradationLevel
	IsLeader() bool
	LeaderNodeID() string
	OnNodeFailure(callback func(nodeID string))
	OnDegradationChanged(callback func(level DegradationLevel))
	ClusterView() ClusterView
}

//Personal.AI order the ending