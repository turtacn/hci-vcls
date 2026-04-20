package e2e

import (
	"testing"

	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/statemachine"
	"github.com/turtacn/hci-vcls/pkg/metrics"
)

func TestStateMachineE2EFlow(t *testing.T) {
	m := metrics.NewNoopMetrics()
	sm := statemachine.NewMachine(m)

	if sm.CurrentLevel() != string(fdm.DegradationNone) {
		t.Errorf("Expected None, got %s", sm.CurrentLevel())
	}

	_ = sm.TransitionString("heartbeat_restored")
	err := sm.TransitionString("degradation_detected")
	if err != nil {
		t.Fatalf("expected nil error on first transition, got %v", err)
	}

	if sm.Current() != statemachine.StateDegraded {
		t.Errorf("Expected degraded, got %s", sm.Current())
	}

	err = sm.TransitionString("evaluation_started")
	if err != nil {
		t.Fatalf("expected nil error on transition, got %v", err)
	}

	if sm.Current() != statemachine.StateEvaluating {
		t.Errorf("Expected evaluating, got %s", sm.Current())
	}

	err = sm.TransitionString("failover_triggered")
	if err != nil {
		t.Fatalf("expected nil error on transition, got %v", err)
	}

	if sm.Current() != statemachine.StateFailover {
		t.Errorf("Expected failover, got %s", sm.Current())
	}
}
