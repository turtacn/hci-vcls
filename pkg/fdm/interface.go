package fdm

import "context"

type Evaluator interface {
	Evaluate(ctx context.Context, clusterID string, leaderNodeID string, hosts []HostState, witnessAvailable bool) (*ClusterState, error)
}

