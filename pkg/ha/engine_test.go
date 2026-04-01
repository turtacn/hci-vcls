package ha

import (
	"context"
	"errors"
	"testing"

	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/qm"
	"github.com/turtacn/hci-vcls/pkg/vcls"
)

type mockEvaluator struct {
	decision HADecision
	err      error
}

func (m *mockEvaluator) Evaluate(ctx context.Context, vmid string, cv ClusterView) (HADecision, error) {
	return m.decision, m.err
}
func (m *mockEvaluator) SelectBootPath(ctx context.Context, vmid string) (BootPath, error) {
	return BootPathMySQL, nil
}
func (m *mockEvaluator) SelectTargetNode(ctx context.Context, vmid string, cv ClusterView) (string, error) {
	return "", nil
}
func (m *mockEvaluator) AssignPriority(ctx context.Context, vmid string) (int, error) {
	return 0, nil
}

type mockBatchExecutor struct {
	err         error
	activeTasks map[string]BootTask
	stats       BatchStats
}

func (m *mockBatchExecutor) Execute(ctx context.Context, tasks []BootTask, policy BatchBootPolicy) error {
	for _, t := range tasks {
		m.activeTasks[t.VMID] = t
	}
	return m.err
}
func (m *mockBatchExecutor) AddTask(task BootTask) error {
	m.activeTasks[task.VMID] = task
	return nil
}
func (m *mockBatchExecutor) CancelTask(vmid string) error {
	delete(m.activeTasks, vmid)
	return nil
}
func (m *mockBatchExecutor) ActiveTasks() map[string]BootTask {
	return m.activeTasks
}
func (m *mockBatchExecutor) Stats() BatchStats {
	return m.stats
}

type mockMySQLAdapter struct {
	claimErr   error
	confirmErr error
	releaseErr error
}

func (m *mockMySQLAdapter) Health() mysql.MySQLStatus {
	return mysql.MySQLStatus{State: mysql.MySQLStateHealthy}
}
func (m *mockMySQLAdapter) ClaimBoot(claim mysql.BootClaim) error            { return m.claimErr }
func (m *mockMySQLAdapter) ConfirmBoot(vmid, token string) error             { return m.confirmErr }
func (m *mockMySQLAdapter) ReleaseBoot(vmid, token string) error             { return m.releaseErr }
func (m *mockMySQLAdapter) GetVMState(vmid string) (*mysql.HAVMState, error) { return nil, nil }
func (m *mockMySQLAdapter) UpsertVMState(state mysql.HAVMState) error        { return nil }
func (m *mockMySQLAdapter) Close() error                                     { return nil }

type mockQMExecutor struct {
	result qm.BootResult
}

func (m *mockQMExecutor) Start(ctx context.Context, vmid string, opts qm.BootOptions) qm.BootResult {
	return m.result
}
func (m *mockQMExecutor) Status(ctx context.Context, vmid string) (qm.VMStatus, error) {
	return qm.VMStatusRunning, nil
}
func (m *mockQMExecutor) Stop(ctx context.Context, vmid string, opts qm.BootOptions) error {
	return nil
}
func (m *mockQMExecutor) Lock(ctx context.Context, vmid string) error   { return nil }
func (m *mockQMExecutor) Unlock(ctx context.Context, vmid string) error { return nil }

type mockVCLSAgent struct{}

func (m *mockVCLSAgent) Start(ctx context.Context) error { return nil }
func (m *mockVCLSAgent) Stop() error                     { return nil }
func (m *mockVCLSAgent) ClusterServiceState() vcls.ClusterServiceState {
	return vcls.ServiceStateHealthy
}
func (m *mockVCLSAgent) IsCapable(cap vcls.Capability) bool                             { return true }
func (m *mockVCLSAgent) RequireCapability(cap vcls.Capability) error                    { return nil }
func (m *mockVCLSAgent) OnDegradationChanged(callback func(level fdm.DegradationLevel)) {}
func (m *mockVCLSAgent) ActiveCapabilities() []vcls.Capability                          { return nil }

func TestEngine_Evaluate(t *testing.T) {
	fdmAgent := &mockFDMAgent{leaderID: "node-1"} // From evaluator_test.go

	evalDecision := HADecision{Action: ActionBoot, TargetNode: "node-2", VMID: "100"}
	evaluator := &mockEvaluator{decision: evalDecision, err: nil}

	engine := NewEngine(
		HAEngineConfig{},
		evaluator,
		&mockBatchExecutor{},
		fdmAgent, // IsLeader() returns true
		&mockMySQLAdapter{},
		&mockQMExecutor{},
		&mockCacheManager{},
		&mockVCLSAgent{},
	)

	ctx := context.Background()

	decision, err := engine.Evaluate(ctx, "100")
	if err != nil {
		t.Fatalf("Expected evaluate to succeed, got %v", err)
	}
	if decision.Action != ActionBoot {
		t.Errorf("Expected ActionBoot, got %v", decision.Action)
	}

	// Test non-leader
	fdmAgentNonLeader := &mockFDMAgent{leaderID: "node-2"}
	_ = NewEngine(
		HAEngineConfig{},
		evaluator,
		&mockBatchExecutor{},
		fdmAgentNonLeader, // IsLeader() returns false for some reason based on logic, let's inject a strict one
		&mockMySQLAdapter{},
		&mockQMExecutor{},
		&mockCacheManager{},
		&mockVCLSAgent{},
	)

	// Override IsLeader to ensure it fails
	// Note: We'd need to change mockFDMAgent to allow overriding IsLeader
}

type mockFDMAgentStrict struct {
	isLeader bool
}

func (m *mockFDMAgentStrict) Start(ctx context.Context) error                                { return nil }
func (m *mockFDMAgentStrict) Stop() error                                                    { return nil }
func (m *mockFDMAgentStrict) NodeStates() map[string]fdm.NodeState                           { return nil }
func (m *mockFDMAgentStrict) LocalDegradationLevel() fdm.DegradationLevel                    { return fdm.DegradationNone }
func (m *mockFDMAgentStrict) IsLeader() bool                                                 { return m.isLeader }
func (m *mockFDMAgentStrict) LeaderNodeID() string                                           { return "" }
func (m *mockFDMAgentStrict) OnNodeFailure(callback func(nodeID string))                     {}
func (m *mockFDMAgentStrict) OnDegradationChanged(callback func(level fdm.DegradationLevel)) {}
func (m *mockFDMAgentStrict) ClusterView() fdm.ClusterView                                   { return fdm.ClusterView{} }

func TestEngine_Evaluate_NotLeader(t *testing.T) {
	evaluator := &mockEvaluator{}
	engine := NewEngine(
		HAEngineConfig{},
		evaluator,
		&mockBatchExecutor{},
		&mockFDMAgentStrict{isLeader: false},
		&mockMySQLAdapter{},
		&mockQMExecutor{},
		&mockCacheManager{},
		&mockVCLSAgent{},
	)

	ctx := context.Background()

	decision, err := engine.Evaluate(ctx, "100")
	if err != ErrNotLeader {
		t.Errorf("Expected ErrNotLeader, got %v", err)
	}
	if decision.Action != ActionSkip {
		t.Errorf("Expected ActionSkip, got %v", decision.Action)
	}
}

func TestEngine_Execute(t *testing.T) {
	mysqlAdp := &mockMySQLAdapter{}
	qmExec := &mockQMExecutor{result: qm.BootResult{Success: true}}

	engine := NewEngine(
		HAEngineConfig{BootTimeoutMs: 1000},
		&mockEvaluator{},
		&mockBatchExecutor{},
		&mockFDMAgentStrict{isLeader: true},
		mysqlAdp,
		qmExec,
		&mockCacheManager{},
		&mockVCLSAgent{},
	)

	ctx := context.Background()
	decision := HADecision{Action: ActionBoot, TargetNode: "node-2", VMID: "100"}

	err := engine.Execute(ctx, decision)
	if err != nil {
		t.Fatalf("Expected execute to succeed, got %v", err)
	}

	// Test claim failure
	mysqlAdp.claimErr = errors.New("claim failed")
	err = engine.Execute(ctx, decision)
	if err == nil {
		t.Errorf("Expected execute to fail on claim error")
	}

	// Test qm start failure
	mysqlAdp.claimErr = nil
	qmExec.result = qm.BootResult{Success: false, Error: errors.New("start failed")}
	err = engine.Execute(ctx, decision)
	if err == nil {
		t.Errorf("Expected execute to fail on qm start error")
	}

	// Test Skip Action
	decisionSkip := HADecision{Action: ActionSkip, TargetNode: "node-2", VMID: "100"}
	err = engine.Execute(ctx, decisionSkip)
	if err != nil {
		t.Errorf("Expected execute to return nil on skip action")
	}
}

func TestEngine_BatchBoot(t *testing.T) {
	batchExec := &mockBatchExecutor{activeTasks: make(map[string]BootTask)}

	engine := NewEngine(
		HAEngineConfig{},
		&mockEvaluator{},
		batchExec,
		&mockFDMAgentStrict{isLeader: true},
		&mockMySQLAdapter{},
		&mockQMExecutor{},
		&mockCacheManager{},
		&mockVCLSAgent{},
	)

	ctx := context.Background()
	decisions := []HADecision{
		{Action: ActionBoot, VMID: "100"},
		{Action: ActionBoot, VMID: "101"},
	}

	err := engine.BatchBoot(ctx, decisions, BatchBootPolicy{MaxConcurrent: 2})
	if err != nil {
		t.Fatalf("Expected BatchBoot to succeed, got %v", err)
	}

	if len(batchExec.activeTasks) != 2 {
		t.Errorf("Expected 2 active tasks, got %d", len(batchExec.activeTasks))
	}

	tasks := engine.ActiveTasks()
	if len(tasks) != 2 {
		t.Errorf("Expected 2 active tasks from engine, got %d", len(tasks))
	}
}

//Personal.AI order the ending
