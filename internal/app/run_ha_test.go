package app

import (
	"context"
	"errors"
	"testing"

	"github.com/turtacn/hci-vcls/internal/config"
	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/vcls"
	"go.uber.org/zap"
)

func TestRunHAOnce_NotLeader(t *testing.T) {
	elector := &mockElector{leader: false}
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, nil, nil, nil, nil, nil, nil, nil)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil)
	if !errors.Is(err, ErrNotLeader) {
		t.Errorf("expected ErrNotLeader, got %v", err)
	}
	if res != nil {
		t.Errorf("expected nil res, got %+v", res)
	}
}

func TestRunHAOnce_CriticalDegradation(t *testing.T) {
	elector := &mockElector{leader: true}
	agent := &mockFDMAgent{level: fdm.DegradationCritical}
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, nil, nil, nil, nil, nil, nil, agent)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil)
	if !errors.Is(err, ErrBelowThreshold) {
		t.Errorf("expected ErrBelowThreshold, got %v", err)
	}
	if res == nil || !res.Skipped {
		t.Errorf("expected result to indicate skipping due to critical degradation")
	}
}

func TestRunHAOnce_NoEligibleVMs(t *testing.T) {
	elector := &mockElector{leader: true}
	agent := &mockFDMAgent{level: fdm.DegradationMajor}
	vclsService := &mockVCLS{eligible: []*vcls.VM{}}
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, nil, nil, nil, nil, nil, agent)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if res == nil || !res.Skipped || res.Reason != "No eligible protected VMs found" {
		t.Errorf("expected skipping due to no VMs, got %+v", res)
	}
}

func TestRunHAOnce_PlannerFails(t *testing.T) {
	elector := &mockElector{leader: true}
	agent := &mockFDMAgent{level: fdm.DegradationMajor}
	vclsService := &mockVCLS{eligible: []*vcls.VM{{ID: "vm-1"}}}
	planner := &mockPlannerErr{}
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, nil, nil, nil, nil, agent)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil)
	if err == nil {
		t.Errorf("expected error from planner")
	}
	if res != nil {
		t.Errorf("expected nil result on error")
	}
}

func TestRunHAOnce_EmptyPlan(t *testing.T) {
	elector := &mockElector{leader: true}
	agent := &mockFDMAgent{level: fdm.DegradationMajor}
	vclsService := &mockVCLS{eligible: []*vcls.VM{{ID: "vm-1"}}}
	planner := &mockPlanner{plan: &ha.Plan{}} // empty tasks
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, nil, nil, nil, nil, agent)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if res == nil || !res.Skipped {
		t.Errorf("expected skipping due to empty plan")
	}
}

func TestRunHAOnce_ExecutorFails(t *testing.T) {
	elector := &mockElector{leader: true}
	agent := &mockFDMAgent{level: fdm.DegradationMajor}
	vclsService := &mockVCLS{eligible: []*vcls.VM{{ID: "vm-1"}}}
	planner := &mockPlanner{plan: &ha.Plan{ID: "plan-1", Tasks: []ha.VMTask{{ID: "task-1"}}}}
	executor := &mockExecutorErr{}
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, executor, nil, nil, nil, agent)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil)
	if err == nil {
		t.Errorf("expected error from executor")
	}
	if res == nil || res.PlanID != "plan-1" {
		t.Errorf("expected result to contain plan ID even if executor fails")
	}
}

func TestRunHAOnce_Success(t *testing.T) {
	elector := &mockElector{leader: true}
	agent := &mockFDMAgent{
		level: fdm.DegradationMinor,
		nodes: map[string]fdm.NodeState{
			"node-1": fdm.NodeStateAlive,
		},
	}
	vclsService := &mockVCLS{eligible: []*vcls.VM{{ID: "vm-1"}}}
	planner := &mockPlanner{plan: &ha.Plan{ID: "plan-2", Tasks: []ha.VMTask{{ID: "task-1"}}}}
	executor := &mockExecutor{}
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, executor, nil, nil, nil, agent)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if res == nil || res.Skipped {
		t.Errorf("expected successful execution")
	}
	if res.PlanID != "plan-2" || res.TaskCount != 1 {
		t.Errorf("expected result to reflect plan, got %+v", res)
	}
}

// Additional mocks for new paths
type mockPlannerErr struct{}

func (m *mockPlannerErr) BuildPlan(ctx context.Context, req ha.PlanRequest) (*ha.Plan, error) {
	return nil, errors.New("planner error")
}

type mockExecutorErr struct{}

func (m *mockExecutorErr) Execute(ctx context.Context, plan *ha.Plan) error {
	return errors.New("executor error")
}
func (m *mockExecutorErr) ExecuteWithCallback(ctx context.Context, plan *ha.Plan, onTaskDone func(ha.VMTask)) error {
	return errors.New("executor error")
}
func (m *mockExecutorErr) ExecuteWithPlan(ctx context.Context, planInterface interface{}) error {
	return errors.New("executor error")
}
