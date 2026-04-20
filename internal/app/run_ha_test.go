package app

import (
	"context"
	"errors"
	"testing"

	"github.com/turtacn/hci-vcls/internal/config"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/vcls"
	"go.uber.org/zap"
)

func TestRunHAOnce_NotLeader(t *testing.T) {
	elector := &mockElector{leader: false}
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil, false)
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
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, nil, nil, nil, nil, nil, nil, agent, nil, nil)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil, false)
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
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, nil, nil, nil, nil, nil, agent, nil, nil)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil, false)
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
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, nil, nil, nil, nil, agent, nil, nil)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil, false)
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
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, nil, nil, nil, nil, agent, nil, nil)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil, false)
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
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, executor, nil, nil, nil, agent, nil, nil)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil, false)
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
	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, executor, nil, nil, nil, agent, nil, nil)

	res, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil, false)
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

func (m *mockExecutorErr) Execute(ctx context.Context, plan *ha.Plan, opts ha.ExecuteOpts) error {
	return errors.New("executor error")
}
func (m *mockExecutorErr) ExecuteWithCallback(ctx context.Context, plan *ha.Plan, opts ha.ExecuteOpts, onTaskDone func(ha.VMTask)) error {
	return errors.New("executor error")
}
func (m *mockExecutorErr) ExecuteWithPlan(ctx context.Context, plan *ha.Plan, opts ha.ExecuteOpts) error {
	return errors.New("executor error")
}

type mockPlanRepo struct {
	createCalls int
	err         error
	lastRecord  *mysql.PlanRecord
}

func (m *mockPlanRepo) Create(ctx context.Context, plan *mysql.PlanRecord) error {
	m.createCalls++
	m.lastRecord = plan
	return m.err
}

func (m *mockPlanRepo) GetByID(ctx context.Context, planID string) (*mysql.PlanRecord, error) {
	return m.lastRecord, nil
}

func TestRunHAOnce_PersistsPlan(t *testing.T) {
	elector := &mockElector{leader: true}
	agent := &mockFDMAgent{level: fdm.DegradationMinor}
	vclsService := &mockVCLS{eligible: []*vcls.VM{{ID: "vm-1"}}}
	planner := &mockPlanner{plan: &ha.Plan{ID: "plan-2", Tasks: []ha.VMTask{{ID: "task-1"}}}}
	executor := &mockExecutor{}
	planRepo := &mockPlanRepo{}

	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, executor, nil, nil, planRepo, agent, nil, nil)

	_, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil, false)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if planRepo.createCalls != 1 {
		t.Errorf("expected planRepo.Create to be called 1 time, got %d", planRepo.createCalls)
	}
	if planRepo.lastRecord == nil || planRepo.lastRecord.ID != "plan-2" {
		t.Errorf("expected plan-2 to be persisted")
	}
}

func TestRunHAOnce_PlanRepoFails_StillExecutes(t *testing.T) {
	elector := &mockElector{leader: true}
	agent := &mockFDMAgent{level: fdm.DegradationMinor}
	vclsService := &mockVCLS{eligible: []*vcls.VM{{ID: "vm-1"}}}
	planner := &mockPlanner{plan: &ha.Plan{ID: "plan-2", Tasks: []ha.VMTask{{ID: "task-1"}}}}
	executor := &mockExecutor{}
	planRepo := &mockPlanRepo{err: errors.New("db down")}

	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, executor, nil, nil, planRepo, agent, nil, nil)

	_, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil, false)
	if err != nil {
		t.Fatalf("expected nil error, execution should proceed even if plan persistence fails, got %v", err)
	}

	if planRepo.createCalls != 1 {
		t.Errorf("expected planRepo.Create to be called")
	}
}

type mockPlanCache struct {
	putErr    error
	deleteErr error
	putCalled bool
	delCalled bool
}

func (m *mockPlanCache) Put(plan *ha.Plan) error {
	m.putCalled = true
	return m.putErr
}

func (m *mockPlanCache) Get(planID string) (*ha.Plan, error) { return nil, nil }
func (m *mockPlanCache) Delete(planID string) error          { m.delCalled = true; return m.deleteErr }
func (m *mockPlanCache) List() ([]*ha.Plan, error)           { return nil, nil }

func TestRunHAOnce_PlanCachePutFails_Aborts(t *testing.T) {
	elector := &mockElector{leader: true}
	agent := &mockFDMAgent{level: fdm.DegradationMinor}
	vclsService := &mockVCLS{eligible: []*vcls.VM{{ID: "vm-1"}}}
	planner := &mockPlanner{plan: &ha.Plan{ID: "plan-cache-1", Tasks: []ha.VMTask{{ID: "task-1"}}}}
	executor := &mockExecutor{}
	planCache := &mockPlanCache{putErr: errors.New("disk full")}

	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, executor, nil, nil, nil, agent, nil, planCache)

	_, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil, false)
	if err == nil {
		t.Errorf("expected error when planCache.Put fails")
	}

	if planCache.putCalled != true {
		t.Errorf("expected planCache.Put to be called")
	}

	if executor.executed.Load() {
		t.Errorf("expected execution to abort when cache put fails")
	}
}

func TestRunHAOnce_Success_ClearsCache(t *testing.T) {
	elector := &mockElector{leader: true}
	agent := &mockFDMAgent{level: fdm.DegradationMinor}
	vclsService := &mockVCLS{eligible: []*vcls.VM{{ID: "vm-1"}}}
	planner := &mockPlanner{plan: &ha.Plan{ID: "plan-cache-2", Tasks: []ha.VMTask{{ID: "task-1"}}}}
	executor := &mockExecutor{}
	planCache := &mockPlanCache{}

	s := NewService(&config.Config{}, zap.NewNop(), nil, elector, nil, vclsService, planner, executor, nil, nil, nil, agent, nil, planCache)

	_, err := s.RunHAOnce(context.Background(), "cluster-1", "auto", nil, false)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if !planCache.putCalled {
		t.Errorf("expected planCache.Put to be called")
	}

	if !planCache.delCalled {
		t.Errorf("expected planCache.Delete to be called after success")
	}
}
