package fdm

import "context"

type AgentConfig struct {
	FDMConfig FDMConfig
}

type DaemonConfig struct {
	Agents []AgentConfig
}

type Daemon interface {
	Start(ctx context.Context) error
	Stop() error
	Agent(id string) (Agent, error)
	Agents() map[string]Agent
	ClusterDegradationLevel() DegradationLevel
	OnAnyNodeFailure(callback func(nodeID string, agentID string))
}

