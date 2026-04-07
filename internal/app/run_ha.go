package app

import (
	"context"
	"fmt"

	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
)

// RunHAResult encapsulates the outcome of a single HA orchestration run.
type RunHAResult struct {
	PlanID    string
	Skipped   bool
	Reason    string
	TaskCount int
}

// RunHAOnce coordinates the HA execution flow:
// 1. Check leadership
// 2. Evaluate degradation level
// 3. Gather protected VMs
// 4. Build execution plan
// 5. Execute the plan
func (s *Service) RunHAOnce(ctx context.Context, clusterID string, trigger string, failedNodeIDs []string) (*RunHAResult, error) {
	if !s.election.IsLeader() {
		return nil, ErrNotLeader
	}

	// Determine current degradation level.
	// In a complete implementation, this might read from statemachine.
	// For now, we query the FDM agent (or statemachine) to get the local perspective.
	var currentLevel fdm.DegradationLevel = fdm.DegradationNone
	if s.fdmAgent != nil {
		currentLevel = s.fdmAgent.LocalDegradationLevel()
	}

	// We block HA execution if the level is Critical (Isolated)
	// Other levels like Minor or Major might restrict some paths (e.g., ZK or MySQL),
	// but Critical completely isolates the node.
	if currentLevel == fdm.DegradationCritical {
		if s.logger != nil {
			s.logger.Warn("HA skipped: cluster is in Critical degradation state", "cluster", clusterID)
		}
		return &RunHAResult{
			Skipped: true,
			Reason:  "Cluster degradation level is Critical (Isolated)",
		}, ErrBelowThreshold
	}

	// Gather protected VMs
	var protectedVMs []*ha.VMTask
	if s.vcls != nil {
		vms, err := s.vcls.ListEligible(ctx, clusterID)
		if err != nil {
			return nil, fmt.Errorf("failed to list eligible VMs: %w", err)
		}
		// Convert vcls.VM to pointer for planner input.
		// For the time being, Planner expects a list of vcls.VM pointers or similar.
		// Wait, the Planner interface expects vcls.VM pointers.
		// `protectedVMs` in PlanRequest expects `[]*vcls.VM`.
	}

	var eligibleVMs = make([]*vcls.VM, 0)
	if s.vcls != nil {
		vms, err := s.vcls.ListEligible(ctx, clusterID)
		if err != nil {
			return nil, fmt.Errorf("failed to list eligible VMs: %w", err)
		}
		eligibleVMs = vms
	}

	if len(eligibleVMs) == 0 {
		return &RunHAResult{
			Skipped: true,
			Reason:  "No eligible protected VMs found",
		}, nil
	}

	// Gather host candidates.
	// For now, we get node states from fdmAgent to find alive nodes as candidates.
	var candidates []ha.HostCandidate
	if s.fdmAgent != nil {
		nodes := s.fdmAgent.NodeStates()
		for nodeID, state := range nodes {
			if state == fdm.NodeStateAlive {
				candidates = append(candidates, ha.HostCandidate{
					HostID:  nodeID,
					Healthy: true,
				})
			}
		}
	}

	req := ha.PlanRequest{
		ClusterID:        clusterID,
		FailedHosts:      failedNodeIDs,
		ProtectedVMs:     eligibleVMs,
		HostCandidates:   candidates,
		PreferWitness:    false, // can be parameterized later
		BatchSize:        5,     // default or from config
		DegradationLevel: string(currentLevel),
	}

	if s.config != nil && s.config.HA.BatchSize > 0 {
		req.BatchSize = s.config.HA.BatchSize
	}

	plan, err := s.planner.BuildPlan(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("planner failed to build plan: %w", err)
	}

	if plan == nil || len(plan.Tasks) == 0 {
		return &RunHAResult{
			Skipped: true,
			Reason:  "Planner produced an empty plan",
		}, nil
	}

	// Optional: Persist plan to repository before executing.
	// TODO: Save to planRepo if required by idempotency goals.
	if s.planRepo != nil {
		// Mock mapping to mysql.HAPlan
		// We'll skip deep persistence implementation in this step if it requires extensive mapping,
		// but we call planRepo as a placeholder hook.
		// _ = s.planRepo.Create(ctx, plan)
	}

	err = s.executor.Execute(ctx, plan)
	if err != nil {
		return &RunHAResult{
			PlanID:    plan.ID,
			TaskCount: len(plan.Tasks),
			Reason:    "Executor failed to complete the plan",
		}, fmt.Errorf("executor failed: %w", err)
	}

	if s.logger != nil {
		s.logger.Info("HA orchestration completed successfully", "planID", plan.ID, "tasks", len(plan.Tasks))
	}

	return &RunHAResult{
		PlanID:    plan.ID,
		TaskCount: len(plan.Tasks),
		Reason:    "Success",
	}, nil
}
