package ha

import (
	"context"

	"github.com/turtacn/hci-vcls/pkg/cache"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/qm"
	"github.com/turtacn/hci-vcls/pkg/vcls"
)

type engineImpl struct {
	config    HAEngineConfig
	evaluator Evaluator
	batchExec BatchExecutor
	fdmAgent  fdm.Agent
	mysqlAdp  mysql.Adapter
	qmExec    qm.Executor
	cacheMgr  cache.CacheManager
	vclsAgent vcls.Agent
	activeMap map[string]BootTask
	cbs       []func(HADecision)
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewEngine(
	config HAEngineConfig,
	evaluator Evaluator,
	batchExec BatchExecutor,
	fdmAgent fdm.Agent,
	mysqlAdp mysql.Adapter,
	qmExec qm.Executor,
	cacheMgr cache.CacheManager,
	vclsAgent vcls.Agent,
) HAEngine {
	ctx, cancel := context.WithCancel(context.Background())
	return &engineImpl{
		config:    config,
		evaluator: evaluator,
		batchExec: batchExec,
		fdmAgent:  fdmAgent,
		mysqlAdp:  mysqlAdp,
		qmExec:    qmExec,
		cacheMgr:  cacheMgr,
		vclsAgent: vclsAgent,
		activeMap: make(map[string]BootTask),
		cbs:       make([]func(HADecision), 0),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (e *engineImpl) Start(ctx context.Context) error {
	// A real implementation might start background tasks
	return nil
}

func (e *engineImpl) Stop() error {
	e.cancel()
	return nil
}

func (e *engineImpl) Evaluate(ctx context.Context, vmid string) (HADecision, error) {
	if !e.fdmAgent.IsLeader() {
		return HADecision{Action: ActionSkip, Reason: "Not the leader node"}, ErrNotLeader
	}

	cv := e.ClusterView()
	decision, err := e.evaluator.Evaluate(ctx, vmid, cv)

	if err == nil {
		for _, cb := range e.cbs {
			cb(decision)
		}
	}

	return decision, err
}

func (e *engineImpl) Execute(ctx context.Context, decision HADecision) error {
	if decision.Action != ActionBoot {
		return nil
	}

	claim := mysql.BootClaim{VMID: decision.VMID, Token: "mock-token", TargetNode: decision.TargetNode}
	err := e.mysqlAdp.ClaimBoot(claim)
	if err != nil {
		return err
	}

	opts := qm.BootOptions{TimeoutMs: e.config.BootTimeoutMs}
	result := e.qmExec.Start(ctx, decision.VMID, opts)

	if result.Success {
		e.mysqlAdp.ConfirmBoot(decision.VMID, claim.Token)
	} else {
		e.mysqlAdp.ReleaseBoot(decision.VMID, claim.Token)
		return result.Error
	}

	return nil
}

func (e *engineImpl) BatchBoot(ctx context.Context, decisions []HADecision, policy BatchBootPolicy) error {
	tasks := make([]BootTask, 0, len(decisions))
	for _, d := range decisions {
		tasks = append(tasks, BootTask{
			VMID:       d.VMID,
			Decision:   d,
			Status:     TaskPending,
			RetryCount: 0,
		})
	}

	return e.batchExec.Execute(ctx, tasks, policy)
}

func (e *engineImpl) ClusterView() ClusterView {
	fdmCV := e.fdmAgent.ClusterView()
	cv := ClusterView{Nodes: make(map[string]fdm.NodeState)}
	for k, v := range fdmCV.Nodes {
		cv.Nodes[k] = v
	}
	return cv
}

func (e *engineImpl) ActiveTasks() map[string]BootTask {
	return e.batchExec.ActiveTasks()
}

func (e *engineImpl) OnDecision(callback func(decision HADecision)) {
	e.cbs = append(e.cbs, callback)
}

//Personal.AI order the ending
