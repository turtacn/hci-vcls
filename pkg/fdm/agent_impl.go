package fdm

import (
	"context"
	"sync"
	"time"

	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/witness"
)

// statemachine and ha references removed here to avoid import cycles.
// In this architecture, components like statemachine and executor should be injected from internal/app via standard callbacks or orchestrator interfaces rather than hard dependencies in fdm.
// Alternatively, we use an interface defined in fdm.

type statemachineWrapper interface {
	TransitionString(event string) error
	EvaluateWithInput(input interface{}) (string, string)
}

type haExecutorWrapper interface {
	ExecuteWithPlan(ctx context.Context, plan interface{}) error
}

type agentImpl struct {
	mu               sync.RWMutex
	config           FDMConfig
	prober           Prober
	elector          election.Elector
	statemachine     statemachineWrapper
	haExecutor       haExecutorWrapper
	witnessClient    witness.Client
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

func (a *agentImpl) SetStateMachine(sm statemachineWrapper) {
	a.statemachine = sm
}

func (a *agentImpl) SetHAExecutor(exe haExecutorWrapper) {
	a.haExecutor = exe
}

func (a *agentImpl) SetWitnessClient(client witness.Client) {
	a.witnessClient = client
}

func (a *agentImpl) Start(ctx context.Context) error {
	a.elector.OnLeaderChange(func(info election.LeaderInfo) {
		a.leaderID = info.NodeID
		a.isLeader = (info.NodeID == a.config.NodeID)
		if a.isLeader {
			if a.metrics != nil {
				a.metrics.IncLeaderChange(a.config.ClusterID)
			}
		}
	})

	err := a.elector.Campaign(a.ctx)
	if err != nil {
		if a.log != nil {
			a.log.Error("Failed to start election campaign", "error", err)
		}
	}

	go a.probeLoop()
	return nil
}

func (a *agentImpl) Stop() error {
	a.cancel()
	return a.elector.Close()
}

func (a *agentImpl) NodeStates() map[string]NodeState {
	a.mu.RLock()
	defer a.mu.RUnlock()

	res := make(map[string]NodeState)
	for k, v := range a.nodeStates {
		res[k] = v
	}
	return res
}

func (a *agentImpl) LocalDegradationLevel() DegradationLevel {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.degradationLevel
}

func (a *agentImpl) IsLeader() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.isLeader
}

func (a *agentImpl) LeaderNodeID() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.leaderID
}

func (a *agentImpl) OnNodeFailure(callback func(nodeID string)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.failureCbs = append(a.failureCbs, callback)
}

func (a *agentImpl) OnDegradationChanged(callback func(level DegradationLevel)) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.degradationCbs = append(a.degradationCbs, callback)
}

func (a *agentImpl) ClusterView() ClusterView {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// Map string level to old int level for compat temporarily
	var oldLevel OldDegradationLevel
	switch a.degradationLevel {
	case DegradationNone:
		oldLevel = OldDegradationNone
	case DegradationMinor:
		oldLevel = OldDegradationZK // mock mapping
	case DegradationMajor:
		oldLevel = OldDegradationCFS
	case DegradationCritical:
		oldLevel = OldDegradationAll
	default:
		oldLevel = OldDegradationNone
	}

	resNodes := make(map[string]NodeState)
	for k, v := range a.nodeStates {
		resNodes[k] = v
	}

	return ClusterView{
		LeaderID:         a.leaderID,
		Nodes:            resNodes,
		DegradationLevel: oldLevel,
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
			finalLevel, reason, baseLevel := a.doProbeRound(a.ctx)
			a.applyDegradation(finalLevel, reason)
			a.maybeTriggerHA(a.ctx, finalLevel, baseLevel)
		}
	}
}

func (a *agentImpl) doProbeRound(ctx context.Context) (DegradationLevel, string, DegradationLevel) {
	res := a.prober.ProbeAll(ctx)
	allFailed := true

	zkState := 2
	cfsState := 2
	mysqlState := 2

	for _, r := range res {
		if r.Success {
			allFailed = false
		}
		if r.Level == HeartbeatL0 {
			if r.Success {
				zkState = 0
				cfsState = 0
				mysqlState = 0
			} else {
				if r.Error != nil && r.Error.Error() == "readonly" {
					zkState = 1
					cfsState = 1
				}
			}
		}
	}

	var baseLevel DegradationLevel
	if allFailed {
		baseLevel = DegradationCritical
	} else {
		baseLevel = DegradationNone
	}

	var finalLevel DegradationLevel = baseLevel
	var reason string

	if a.statemachine != nil {
		input := map[string]interface{}{
			"ZKState":    zkState,
			"CFSState":   cfsState,
			"MySQLState": mysqlState,
			"FDMLevel":   baseLevel,
		}
		levelStr, r := a.statemachine.EvaluateWithInput(input)
		finalLevel = DegradationLevel(levelStr)
		reason = r

		if a.log != nil && finalLevel != DegradationNone {
			a.log.Info("State machine evaluation resulted in degradation", "level", finalLevel, "reason", reason)
		}

		if allFailed || finalLevel != DegradationNone {
			_ = a.statemachine.TransitionString("degradation_detected")
		} else {
			_ = a.statemachine.TransitionString("heartbeat_restored")
		}
	}

	return finalLevel, reason, baseLevel
}

func (a *agentImpl) applyDegradation(finalLevel DegradationLevel, reason string) {
	var changed bool
	a.mu.Lock()
	if finalLevel != a.degradationLevel {
		a.degradationLevel = finalLevel
		changed = true
	}
	a.mu.Unlock()

	if changed {
		if a.metrics != nil {
			a.metrics.SetDegradationLevel(a.config.ClusterID, float64(LevelWeight(finalLevel)))
		}

		a.mu.RLock()
		cbs := make([]func(DegradationLevel), len(a.degradationCbs))
		copy(cbs, a.degradationCbs)
		a.mu.RUnlock()

		for _, cb := range cbs {
			cb(finalLevel)
		}
	}
}

func (a *agentImpl) maybeTriggerHA(ctx context.Context, finalLevel DegradationLevel, baseLevel DegradationLevel) {
	a.mu.RLock()
	isLeader := a.isLeader
	a.mu.RUnlock()

	if isLeader && a.haExecutor != nil && baseLevel == DegradationCritical && finalLevel != DegradationCritical {
		// Double check with witness if cluster size is 2
		if len(a.config.HeartbeatPeers) == 1 && a.witnessClient != nil {
			confirmed, err := a.witnessClient.ConfirmNodeFailure(ctx, a.config.HeartbeatPeers[0])
			if err != nil || !confirmed {
				if a.log != nil {
					a.log.Warn("Witness did not confirm node failure in 2-node cluster, skipping HA", "node", a.config.HeartbeatPeers[0])
				}
				return
			}
		}

		if a.statemachine != nil {
			_ = a.statemachine.TransitionString("evaluation_started")
			_ = a.statemachine.TransitionString("failover_triggered")
		}
		if a.log != nil {
			a.log.Info("Leader detected node failure but cluster is not isolated, triggering HA loop", "node", a.config.NodeID)
		}
		_ = a.haExecutor.ExecuteWithPlan(ctx, nil)
	}
}
