package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/vcls"
	"github.com/turtacn/hci-vcls/test/e2e/helpers"
)

func TestBatchBootLimits(t *testing.T) {
	cfg := helpers.DefaultTestConfig()
	cfg.HA.BatchSize = 3 // small batch size
	testApp := helpers.NewTestService(cfg)

	ctx := context.Background()

	var vms []*vcls.VM
	for i := 1; i <= 10; i++ {
		vmid := fmt.Sprintf("10%d", i)
		_ = testApp.Repo.Upsert(ctx, &mysql.VMRecord{
			VMID:        vmid,
			ClusterID:   cfg.Node.ClusterID,
			Protected:   true,
			PowerState:  "running",
			CurrentHost: "node-failed",
		})
		vms = append(vms, &vcls.VM{
			ID:            vmid,
			ClusterID:     cfg.Node.ClusterID,
			CurrentHost:   "node-failed",
			EligibleForHA: true,
		})
	}

	candidates := []ha.HostCandidate{
		{HostID: "node-good", Healthy: true},
	}
	evalCtx := context.WithValue(ctx, "mock_candidates", candidates)

	req := ha.PlanRequest{
		ClusterID:        cfg.Node.ClusterID,
		FailedHosts:      []string{"node-failed"},
		ProtectedVMs:     vms,
		HostCandidates:   candidates,
		DegradationLevel: "Major",
		BatchSize:        3,
	}

	plan, err := testApp.Planner.BuildPlan(evalCtx, req)
	if err != nil {
		t.Fatalf("BuildPlan failed: %v", err)
	}

	if plan.TotalBatches != 4 {
		t.Errorf("Expected 4 batches for 10 VMs with size 3, got %d", plan.TotalBatches)
	}

	// Manually inject records simulating the task creation step usually in EvaluateHA
	for _, task := range plan.Tasks {
		_ = testApp.TaskRepo.Create(ctx, &mysql.HATaskRecord{
			ID:        task.ID,
			VMID:      task.VMID,
			ClusterID: task.ClusterID,
			Status:    mysql.TaskPending,
		})
	}

	// Just execute the first batch or let it run
	err = testApp.Executor.Execute(evalCtx, plan, ha.ExecuteOpts{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	tasks, _ := testApp.TaskRepo.ListByPlan(ctx, plan.ID)
	completed := 0
	for _, task := range tasks {
		if task.Status == mysql.TaskCompleted {
			completed++
		}
	}
	if completed != 10 {
		t.Errorf("Expected 10 tasks to complete natively, got %d", completed)
	}
}
