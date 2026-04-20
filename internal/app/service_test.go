package app

import (
	"context"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/config"
	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/statemachine"
	"github.com/turtacn/hci-vcls/pkg/vcls"
	"go.uber.org/zap"
	"sync/atomic"
)

type mockElector struct {
	leader bool
}

func (m *mockElector) IsLeader() bool { return m.leader }
func (m *mockElector) Status() election.LeaderStatus {
	return election.LeaderStatus{LeaderID: "node1"}
}
func (m *mockElector) Campaign(ctx context.Context) error                     { return nil }
func (m *mockElector) Resign(ctx context.Context) error                       { return nil }
func (m *mockElector) Watch() <-chan election.LeaderStatus                    { return nil }
func (m *mockElector) Close() error                                           { return nil }
func (m *mockElector) OnLeaderChange(callback func(info election.LeaderInfo)) {}
func (m *mockElector) ReceivePeerState(peerNodeID string, peerTerm int64, peerVoteFor string, isLeader bool) {
}
func (m *mockElector) CurrentTermAndVote() (int64, string, bool) { return 0, "", false }
func (m *mockElector) SetNodesCount(count int)                   {}

type mockVCLS struct {
	eligible []*vcls.VM
}

func (m *mockVCLS) Refresh(ctx context.Context, clusterID string) error      { return nil }
func (m *mockVCLS) GetVM(ctx context.Context, vmID string) (*vcls.VM, error) { return nil, nil }
func (m *mockVCLS) ListProtected(ctx context.Context, clusterID string) ([]*vcls.VM, error) {
	return nil, nil
}
func (m *mockVCLS) ListEligible(ctx context.Context, clusterID string) ([]*vcls.VM, error) {
	return m.eligible, nil
}

type mockPlanner struct {
	plan *ha.Plan
}

func (m *mockPlanner) BuildPlan(ctx context.Context, req ha.PlanRequest) (*ha.Plan, error) {
	return m.plan, nil
}

type mockExecutor struct {
	executed atomic.Bool
}

func (m *mockExecutor) Execute(ctx context.Context, plan *ha.Plan, opts ha.ExecuteOpts) error {
	return nil
}
func (m *mockExecutor) ExecuteWithCallback(ctx context.Context, plan *ha.Plan, opts ha.ExecuteOpts, cb func(ha.VMTask)) error {
	m.executed.Store(true)
	return nil
}

func (m *mockExecutor) ExecuteWithPlan(ctx context.Context, plan *ha.Plan, opts ha.ExecuteOpts) error {
	m.executed.Store(true)
	return nil
}

type mockFDMAgent struct {
	level fdm.DegradationLevel
	nodes map[string]fdm.NodeState
}

func (m *mockFDMAgent) Start(ctx context.Context) error                 { return nil }
func (m *mockFDMAgent) Stop() error                                     { return nil }
func (m *mockFDMAgent) LocalDegradationLevel() fdm.DegradationLevel     { return m.level }
func (m *mockFDMAgent) OnDegradationChanged(func(fdm.DegradationLevel)) {}
func (m *mockFDMAgent) OnNodeFailure(func(string))                      {}
func (m *mockFDMAgent) NodeStates() map[string]fdm.NodeState {
	if m.nodes != nil {
		return m.nodes
	}
	return map[string]fdm.NodeState{}
}
func (m *mockFDMAgent) IsLeader() bool               { return true }
func (m *mockFDMAgent) LeaderNodeID() string         { return "node1" }
func (m *mockFDMAgent) ClusterView() fdm.ClusterView { return fdm.ClusterView{} }

func TestEvaluateHA_NotLeader(t *testing.T) {
	elector := &mockElector{leader: false}
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	_, err := s.EvaluateHA(context.Background(), "c1")
	if err != ErrNotLeader {
		t.Errorf("Expected ErrNotLeader, got %v", err)
	}
}

func TestEvaluateHA_BelowThreshold(t *testing.T) {
	elector := &mockElector{leader: true}
	fdmAgent := &mockFDMAgent{level: fdm.DegradationMinor}
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, nil, nil, nil, nil, nil, nil, fdmAgent, nil, nil)

	_, err := s.EvaluateHA(context.Background(), "c1")
	if err != ErrBelowThreshold {
		t.Errorf("Expected ErrBelowThreshold, got %v", err)
	}
}

func TestEvaluateHA_EmptyCluster(t *testing.T) {
	elector := &mockElector{leader: true}
	fdmAgent := &mockFDMAgent{level: fdm.DegradationMajor}
	vclsService := &mockVCLS{eligible: []*vcls.VM{}}

	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, nil, nil, nil, nil, nil, fdmAgent, nil, nil)

	plan, err := s.EvaluateHA(context.Background(), "c1")
	if err != nil {
		t.Errorf("Expected nil error for empty cluster, got %v", err)
	}
	if plan == nil || len(plan.Tasks) > 0 {
		t.Errorf("Expected empty plan, got %v", plan)
	}
}

func TestEvaluateHA_NormalPath(t *testing.T) {
	elector := &mockElector{leader: true}
	fdmAgent := &mockFDMAgent{level: fdm.DegradationMajor}

	vms := []*vcls.VM{
		{ID: "vm1", ClusterID: "c1"},
	}
	vclsService := &mockVCLS{eligible: vms}

	expectedPlan := &ha.Plan{
		ID: "p1", ClusterID: "c1", Tasks: []ha.VMTask{{ID: "t1"}},
	}
	planner := &mockPlanner{plan: expectedPlan}

	executor := &mockExecutor{}
	sm := statemachine.NewMachine(nil)
	_ = sm.Transition(statemachine.EventHeartbeatRestored) // to Stable
	_ = sm.Transition(statemachine.EventHeartbeatLost)     // to Degraded
	_ = sm.Transition(statemachine.EventEvaluationStarted) // to Evaluating

	planRepo := mysql.NewMemoryPlanRepository()

	cfg := &config.Config{HA: config.HAConfig{BatchSize: 5}}

	s := NewService(cfg, zap.NewNop(), nil, elector, nil, vclsService, planner, executor, sm, nil, planRepo, fdmAgent, nil, nil)

	plan, err := s.EvaluateHA(context.Background(), "c1")
	if err != nil {
		t.Fatalf("EvaluateHA failed: %v", err)
	}

	if plan.ID != "p1" {
		t.Errorf("Expected plan p1, got %v", plan.ID)
	}

	if sm.Current() != statemachine.StateFailover {
		t.Errorf("Expected state Failover, got %v", sm.Current())
	}

	// Check repo
	dbPlan, _ := planRepo.GetByID(context.Background(), "p1")
	if dbPlan == nil {
		t.Errorf("Expected plan saved to repo")
	}

	// Need to wait slightly for async executor call to register
	time.Sleep(10 * time.Millisecond)
	if !executor.executed.Load() {
		t.Errorf("Expected executor to be called")
	}
}
