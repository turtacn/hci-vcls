package fdm

import (
	"context"
	"time"

	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
)

type agentImpl struct {
	config           FDMConfig
	prober           Prober
	elector          election.Elector
	log              logger.Logger
	metrics          metrics.Metrics
	nodeStates       map[string]NodeState
	degradationLevel DegradationLevel
	leaderID         string
	isLeader         bool
	failureCbs       []func(string)
	degradationCbs   []func(DegradationLevel)
	ctx              context.Context
	cancel           context.CancelFunc
}

func NewAgent(config FDMConfig, prober Prober, elector election.Elector, log logger.Logger, m metrics.Metrics) Agent {
	ctx, cancel := context.WithCancel(context.Background())
	agent := &agentImpl{
		config:           config,
		prober:           prober,
		elector:          elector,
		log:              log,
		metrics:          m,
		nodeStates:       make(map[string]NodeState),
		degradationLevel: DegradationNone,
		failureCbs:       make([]func(string), 0),
		degradationCbs:   make([]func(DegradationLevel), 0),
		ctx:              ctx,
		cancel:           cancel,
	}

	for _, peer := range config.HeartbeatPeers {
		agent.nodeStates[peer] = NodeStateAlive
	}

	return agent
}

func (a *agentImpl) Start(ctx context.Context) error {
	a.elector.OnLeaderChange(func(info election.LeaderInfo) {
		a.leaderID = info.NodeID
		a.isLeader = (info.NodeID == a.config.NodeID)
		if a.isLeader {
			a.metrics.IncLeaderChanges(metrics.MetricLabels{metrics.LabelClusterID: a.config.ClusterID, metrics.LabelNodeID: a.config.NodeID})
		}
	})

	err := a.elector.Campaign(a.ctx)
	if err != nil {
		a.log.Error("Failed to start election campaign", "error", err)
	}

	go a.probeLoop()
	return nil
}

func (a *agentImpl) Stop() error {
	a.cancel()
	return a.elector.Close()
}

func (a *agentImpl) NodeStates() map[string]NodeState {
	return a.nodeStates
}

func (a *agentImpl) LocalDegradationLevel() DegradationLevel {
	return a.degradationLevel
}

func (a *agentImpl) IsLeader() bool {
	return a.isLeader
}

func (a *agentImpl) LeaderNodeID() string {
	return a.leaderID
}

func (a *agentImpl) OnNodeFailure(callback func(nodeID string)) {
	a.failureCbs = append(a.failureCbs, callback)
}

func (a *agentImpl) OnDegradationChanged(callback func(level DegradationLevel)) {
	a.degradationCbs = append(a.degradationCbs, callback)
}

func (a *agentImpl) ClusterView() ClusterView {
	return ClusterView{
		LeaderID:         a.leaderID,
		Nodes:            a.nodeStates,
		DegradationLevel: a.degradationLevel,
	}
}

func (a *agentImpl) probeLoop() {
	ticker := time.NewTicker(time.Duration(a.config.ProbeIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			res := a.prober.ProbeAll(a.ctx)
			allFailed := true
			for _, r := range res {
				if r.Success {
					allFailed = false
				}
			}

			var newLevel DegradationLevel
			if allFailed {
				newLevel = DegradationAll
			} else {
				newLevel = DegradationNone
			}

			if newLevel != a.degradationLevel {
				a.degradationLevel = newLevel
				a.metrics.SetDegradationLevel(float64(newLevel), metrics.MetricLabels{metrics.LabelClusterID: a.config.ClusterID})
				for _, cb := range a.degradationCbs {
					cb(newLevel)
				}
			}
		}
	}
}

//Personal.AI order the ending