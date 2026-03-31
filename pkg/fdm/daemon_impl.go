package fdm

import (
	"context"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
)

type daemonImpl struct {
	config DaemonConfig
	agents map[string]Agent
	log    logger.Logger
	m      metrics.Metrics
	cbs    []func(string, string)
}

func NewDaemon(config DaemonConfig, log logger.Logger, m metrics.Metrics) Daemon {
	return &daemonImpl{
		config: config,
		agents: make(map[string]Agent),
		log:    log,
		m:      m,
		cbs:    make([]func(string, string), 0),
	}
}

func (d *daemonImpl) Start(ctx context.Context) error {
	for _, aCfg := range d.config.Agents {
		// In a real scenario, dependencies would be injected or constructed here
		// For the skeleton, we mock agent creation or assume agents are added externally
		d.log.Info("Starting agent", "id", aCfg.FDMConfig.NodeID)
	}
	return nil
}

func (d *daemonImpl) Stop() error {
	for id, agent := range d.agents {
		err := agent.Stop()
		if err != nil {
			d.log.Error("Failed to stop agent", "id", id, "error", err)
		}
	}
	return nil
}

func (d *daemonImpl) Agent(id string) (Agent, error) {
	agent, ok := d.agents[id]
	if !ok {
		return nil, ErrNodeNotFound
	}
	return agent, nil
}

func (d *daemonImpl) Agents() map[string]Agent {
	return d.agents
}

func (d *daemonImpl) ClusterDegradationLevel() DegradationLevel {
	// Aggregate degradation levels from all agents
	var maxLevel DegradationLevel = DegradationNone
	for _, agent := range d.agents {
		level := agent.LocalDegradationLevel()
		if level > maxLevel {
			maxLevel = level
		}
	}
	return maxLevel
}

func (d *daemonImpl) OnAnyNodeFailure(callback func(nodeID string, agentID string)) {
	d.cbs = append(d.cbs, callback)
}

//Personal.AI order the ending