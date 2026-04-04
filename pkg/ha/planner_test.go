package ha

import (
	"context"
	"testing"

	"github.com/turtacn/hci-vcls/pkg/vcls"
)

func TestPlanner_BuildPlan(t *testing.T) {
	planner := NewPlanner()
	ctx := context.Background()

	// Test missing cluster ID
	_, err := planner.BuildPlan(ctx, PlanRequest{})
	if err != ErrEmptyClusterID {
		t.Errorf("Expected ErrEmptyClusterID, got %v", err)
	}

	// Test missing protected VMs
	_, err = planner.BuildPlan(ctx, PlanRequest{
		ClusterID: "cluster-1",
	})
	if err != ErrNoProtectedVMs {
		t.Errorf("Expected ErrNoProtectedVMs, got %v", err)
	}

	// Test missing host candidates
	_, err = planner.BuildPlan(ctx, PlanRequest{
		ClusterID: "cluster-1",
		ProtectedVMs: []*vcls.VM{
			{ID: "vm-1", EligibleForHA: true, CurrentHost: "host-1"},
		},
	})
	if err != ErrNoCandidateHost {
		t.Errorf("Expected ErrNoCandidateHost, got %v", err)
	}

	// Test valid request
	plan, err := planner.BuildPlan(ctx, PlanRequest{
		ClusterID: "cluster-1",
		ProtectedVMs: []*vcls.VM{
			{ID: "vm-1", EligibleForHA: true, CurrentHost: "host-1"},
		},
		FailedHosts: []string{"host-1"},
		HostCandidates: []HostCandidate{
			{HostID: "host-2", Healthy: true},
		},
	})
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}
	if plan == nil {
		t.Fatalf("Expected valid plan, got nil")
	}
	if len(plan.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(plan.Tasks))
	}
}

func TestScoreHost(t *testing.T) {
	score, _ := ScoreHost(HostCandidate{HostID: "host-1", Healthy: false}, "", false)
	if score != 0 {
		t.Errorf("Expected 0 score for unhealthy host, got %v", score)
	}

	score, _ = ScoreHost(HostCandidate{HostID: "host-1", Healthy: true, CurrentLoad: 0, FaultDomain: "A", RecentFailures: 0, WitnessCapable: true}, "B", true)
	if score <= 100 {
		t.Errorf("Expected score > 100 for ideal host, got %v", score)
	}
}

// Personal.AI order the ending
