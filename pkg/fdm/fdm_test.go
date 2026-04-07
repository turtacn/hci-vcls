package fdm

import (
	"context"
	"testing"
)

func TestEvaluatorImpl(t *testing.T) {
	eval := NewEvaluator()
	ctx := context.Background()

	// 1. Empty hosts
	state, err := eval.Evaluate(ctx, "cluster1", "node1", []HostState{}, true)
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if state.Degradation != DegradationNone {
		t.Errorf("expected none degradation, got %v", state.Degradation)
	}

	// 2. All healthy
	hosts := []HostState{
		{NodeID: "node1", Healthy: true},
		{NodeID: "node2", Healthy: true},
		{NodeID: "node3", Healthy: true},
	}
	state, _ = eval.Evaluate(ctx, "cluster1", "node1", hosts, true)
	if state.Degradation != DegradationNone {
		t.Errorf("expected none degradation, got %v", state.Degradation)
	}

	// 3. 1 unhealthy (Minor vs Major)
	// v1 rule (1 unhealthy) -> Minor
	// v2 rule (ratio = 1/3 = 0.33) -> Major
	// Max(Minor, Major) -> Major
	hosts[1].Healthy = false
	state, _ = eval.Evaluate(ctx, "cluster1", "node1", hosts, true)
	if state.Degradation != DegradationMajor {
		t.Errorf("expected Major degradation (v2 rule overrides), got %v", state.Degradation)
	}

	// 4. 2 unhealthy, leader at risk (Critical)
	hosts[0].Healthy = false
	state, _ = eval.Evaluate(ctx, "cluster1", "node1", hosts, true)
	if state.Degradation != DegradationCritical {
		t.Errorf("expected critical degradation due to leader risk, got %v", state.Degradation)
	}
	if !state.LeaderAtRisk {
		t.Error("expected LeaderAtRisk to be true")
	}

	// 5. Quorum risk (Critical)
	hosts = []HostState{
		{NodeID: "node1", Healthy: true},
		{NodeID: "node2", Healthy: false},
		{NodeID: "node3", Healthy: false},
	}
	state, _ = eval.Evaluate(ctx, "cluster1", "node1", hosts, true)
	if state.Degradation != DegradationCritical {
		t.Errorf("expected critical degradation due to quorum risk, got %v", state.Degradation)
	}
	if !state.QuorumRisk {
		t.Error("expected QuorumRisk to be true")
	}

	// 6. Witness missing
	hosts = []HostState{
		{NodeID: "node1", Healthy: true},
		{NodeID: "node2", Healthy: true},
		{NodeID: "node3", Healthy: false},
	}
	state, _ = eval.Evaluate(ctx, "cluster1", "node1", hosts, false) // 1 unhealthy, witness false
	if state.Degradation != DegradationMajor { // upgraded to Major due to witness missing
		t.Errorf("expected major degradation due to missing witness, got %v", state.Degradation)
	}
}

func TestEvaluatorImpl_CtxCanceled(t *testing.T) {
	eval := NewEvaluator()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel
	_, err := eval.Evaluate(ctx, "cluster1", "node1", []HostState{}, true)
	if err == nil {
		t.Error("expected context error, got nil")
	}
}
