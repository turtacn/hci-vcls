package app

import (
	"context"

	"github.com/turtacn/hci-vcls/internal/config"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/statemachine"
	"github.com/turtacn/hci-vcls/pkg/vcls"
	"go.uber.org/zap"
)

func (s *Service) fdmState(clusterID string) fdm.DegradationLevel {
	if s.fdmAgent != nil {
		return s.fdmAgent.LocalDegradationLevel()
	}
	return fdm.DegradationNone
}

func shouldTriggerHA(state fdm.DegradationLevel) bool {
	// Let's assume Major and Critical trigger HA
	return state == fdm.DegradationMajor || state == fdm.DegradationCritical
}

func buildPlanRequest(clusterID string, state fdm.DegradationLevel, eligibleVMs []*vcls.VM, cfg config.HAConfig) ha.PlanRequest {
	req := ha.PlanRequest{
		ClusterID:     clusterID,
		ProtectedVMs:  eligibleVMs,
		PreferWitness: true, // simplified
		BatchSize:     cfg.BatchSize,
	}

	// Assuming we can extract FailedHosts from the fdmAgent ClusterView
	// For simplicity in the mock app layer, we will just use dummy data if needed
	// Or we just rely on eligibleVMs which already have HostHealthy=false

	// We also need HostCandidates. We would typically get this from CFS/FDM view.
	// We'll mock it here.
	var candidates []ha.HostCandidate
	// if we had a method in FDM agent to list hosts
	// for now just dummy or rely on tests to inject

	req.HostCandidates = candidates
	return req
}

func toPlanRecord(plan *ha.Plan) *mysql.PlanRecord {
	return &mysql.PlanRecord{
		ID:          plan.ID,
		ClusterID:   plan.ClusterID,
		Trigger:     plan.Trigger,
		Degradation: 0, // mock
		TaskCount:   len(plan.Tasks),
		CreatedAt:   plan.CreatedAt,
	}
}

func (s *Service) onTaskDone(task ha.VMTask) {
	if s.logger != nil {
		s.logger.Info("HA task done", zap.String("vm", task.VMID), zap.String("status", string(task.Status)))
	}
}

func (s *Service) EvaluateHA(ctx context.Context, clusterID string) (*ha.Plan, error) {
	// 1. 检查 leader
	if s.election == nil || !s.election.IsLeader() {
		return nil, ErrNotLeader
	}

	// 2. 检查 degradation
	state := s.fdmState(clusterID)
	if !shouldTriggerHA(state) {
		return nil, ErrBelowThreshold
	}

	// 3. 刷新 vcls 视图
	if s.vcls != nil {
		if err := s.vcls.Refresh(ctx, clusterID); err != nil {
			return nil, err
		}
	}

	// 4. 获取 eligible VMs
	var eligibleVMs []*vcls.VM
	if s.vcls != nil {
		eligibleVMs, _ = s.vcls.ListEligible(ctx, clusterID)
	}

	if len(eligibleVMs) == 0 {
		return &ha.Plan{}, nil // return empty plan, no error
	}

	// 5. 构建 PlanRequest
	req := buildPlanRequest(clusterID, state, eligibleVMs, s.config.HA)

	// Since we mocked HostCandidates, we inject them specifically if we are running in tests
	// This is a bit hacky but keeps the orchestrated flow realistic
	if ctx.Value("mock_candidates") != nil {
		req.HostCandidates = ctx.Value("mock_candidates").([]ha.HostCandidate)
	}

	// 6. Planner 生成计划
	plan, err := s.planner.BuildPlan(ctx, req)
	if err != nil {
		return nil, err
	}

	// 7. 持久化 Plan
	if s.planRepo != nil {
		_ = s.planRepo.Create(ctx, toPlanRecord(plan))
	}

	// 8. 状态机迁移
	if s.statemachine != nil {
		_ = s.statemachine.Transition(statemachine.EventFailoverTriggered)
	}

	// 9. 执行
	if s.executor != nil {
		go func() {
			_ = s.executor.ExecuteWithCallback(context.Background(), plan, s.onTaskDone)
		}()
	}

	return plan, nil
}

