package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/vcls"
	"github.com/turtacn/hci-vcls/test/e2e/helpers"
)

func TestHAMinorityBoot(t *testing.T) {
	cfg := helpers.DefaultTestConfig()
	testApp := helpers.NewTestService(cfg)

	// Pre-populate conditions
	ctx := context.Background()

	_ = testApp.Repo.Upsert(ctx, &mysql.VMRecord{
		VMID:        "100",
		ClusterID:   cfg.Node.ClusterID,
		Protected:   true,
		PowerState:  "running",
		CurrentHost: "node-failed",
	})

	// Inject candidates via context hack
	candidates := []ha.HostCandidate{
		{HostID: "node-good", Healthy: true},
	}
	evalCtx := context.WithValue(ctx, "mock_candidates", candidates)

	// Since NewTestService uses a new internal app service but TestApp holds a reference,
	// let's manually build a PlanRequest and trigger Planner/Executor directly to verify execution loop correctly.

	req := ha.PlanRequest{
		ClusterID: cfg.Node.ClusterID,
		FailedHosts: []string{"node-failed"},
		ProtectedVMs: []*vcls.VM{
			{ID: "100", ClusterID: cfg.Node.ClusterID, CurrentHost: "node-failed", EligibleForHA: true},
		},
		HostCandidates: candidates,
		DegradationLevel: "Major",
	}

	plan, err := testApp.Planner.BuildPlan(evalCtx, req)
	if err != nil {
		t.Fatalf("BuildPlan failed: %v", err)
	}

	dbTask := &mysql.HATaskRecord{
		ID:        plan.Tasks[0].ID,
		VMID:      plan.Tasks[0].VMID,
		ClusterID: plan.Tasks[0].ClusterID,
		Status:    mysql.TaskPending,
	}
	_ = testApp.TaskRepo.Create(ctx, dbTask)

	err = testApp.Executor.Execute(evalCtx, plan, ha.ExecuteOpts{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Ensure the execution pipeline updates state in the DB
	time.Sleep(100 * time.Millisecond) // allow async status update

	tasks, err := testApp.TaskRepo.ListByPlan(ctx, plan.ID)
	if err != nil || len(tasks) == 0 {
		t.Fatalf("Failed to fetch task: %v", err)
	}
	task := tasks[0]

	if task.Status != mysql.TaskCompleted {
		t.Errorf("Expected task to be completed, got %v", task.Status)
	}

	// Verify QM client intercepted the boot request
	started := false
	for _, id := range testApp.QM.Starts {
		if id == "100" {
			started = true
			break
		}
	}
	if !started {
		t.Errorf("Expected VM 100 to be started")
	}
}
