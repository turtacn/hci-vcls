package e2e

import (
	"context"
	"testing"

	"github.com/turtacn/hci-vcls/pkg/fdm"
)

func TestFDMEvaluationCriticalIsolation(t *testing.T) {
	evaluator := fdm.NewEvaluator()

	ctx := context.Background()

	hosts := []fdm.HostState{
		{NodeID: "node-1", Healthy: true},
		{NodeID: "node-2", Healthy: false},
		{NodeID: "node-3", Healthy: false},
	}

	state, err := evaluator.Evaluate(ctx, "cluster-1", "node-1", hosts, true)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if state.Degradation != fdm.DegradationCritical {
		t.Errorf("Expected Critical isolation due to losing all peers natively, got %s", state.Degradation)
	}
}
