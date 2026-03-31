package statemachine

import (
	"context"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/zk"
)

type mockProber struct {
	input EvaluationInput
}

func (m *mockProber) Sample(ctx context.Context) EvaluationInput {
	return m.input
}

func TestMachine_ForceEvaluate(t *testing.T) {
	config := StateMachineConfig{
		EvaluationIntervalMs: 100,
		CooldownMs:           0, // No cooldown for easy testing
	}

	prober := &mockProber{
		input: EvaluationInput{
			ZKStatus: zk.ZKStatus{State: zk.ZKStateUnavailable},
		},
	}
	machine := NewMachine(config, &HealthProber{}).(*machineImpl)
	machine.prober = &HealthProber{} // Hack: Replace with actual mock prober interface if defined. In this code, prober is a concrete struct *HealthProber.
	// Since HealthProber is a concrete type, we'll test the logic directly or recreate it.

	// Better approach for testing machine logic without mocking the concrete prober:
	ctx := context.Background()

	// Direct test of Evaluate function
	res := Evaluate(prober.input)
	if res.Level != fdm.DegradationZK {
		t.Errorf("Expected DegradationZK, got %v", res.Level)
	}

	// Test machine with a real prober but mocked adapters
	hp := NewHealthProber(&mockZKAdapter{state: zk.ZKStateUnavailable}, nil, nil, nil)
	m := NewMachine(config, hp)

	// Force Evaluate
	result, err := m.ForceEvaluate(ctx)
	if err != nil {
		t.Fatalf("Expected ForceEvaluate to succeed, got %v", err)
	}

	if result.Level != fdm.DegradationZK {
		t.Errorf("Expected result level to be DegradationZK, got %v", result.Level)
	}

	if m.CurrentLevel() != fdm.DegradationZK {
		t.Errorf("Expected machine current level to be DegradationZK")
	}

	history := m.TransitionHistory()
	if len(history) != 1 {
		t.Errorf("Expected 1 transition in history, got %d", len(history))
	}

	last := m.LastTransition()
	if last.To != fdm.DegradationZK {
		t.Errorf("Expected last transition to be DegradationZK")
	}

	// Test cooldown
	m2 := NewMachine(StateMachineConfig{CooldownMs: 1000}, hp)
	_, _ = m2.ForceEvaluate(ctx) // First evaluate sets lastEvalAt

	_, err = m2.ForceEvaluate(ctx) // Immediate second evaluate should fail
	if err != ErrCooldownActive {
		t.Errorf("Expected ErrCooldownActive, got %v", err)
	}
}

func TestMachine_StartStop(t *testing.T) {
	config := StateMachineConfig{
		EvaluationIntervalMs: 10,
		CooldownMs:           0,
	}

	hp := NewHealthProber(&mockZKAdapter{state: zk.ZKStateHealthy}, nil, nil, nil)
	m := NewMachine(config, hp)
	ctx := context.Background()

	err := m.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed")
	}

	// Change state
	hp.zkAdapter = &mockZKAdapter{state: zk.ZKStateUnavailable}

	// Wait for loop to pick it up
	time.Sleep(50 * time.Millisecond)

	if m.CurrentLevel() != fdm.DegradationZK {
		t.Errorf("Expected machine current level to be DegradationZK, got %v", m.CurrentLevel())
	}

	err = m.Stop()
	if err != nil {
		t.Fatalf("Expected stop to succeed")
	}
}

func TestTransitionError(t *testing.T) {
	err := ErrMachineNotStarted
	if err.Error() == "" {
		t.Errorf("Expected error message to not be empty")
	}
}

// Helper mock
type mockZKAdapter struct {
	state zk.HealthState
}
func (m *mockZKAdapter) Health() zk.ZKStatus       { return zk.ZKStatus{State: m.state} }
func (m *mockZKAdapter) IsReadOnly() zk.ZKStatus   { return zk.ZKStatus{State: zk.ZKStateHealthy} }
func (m *mockZKAdapter) Ping() zk.ZKStatus         { return zk.ZKStatus{State: zk.ZKStateHealthy} }
func (m *mockZKAdapter) Close() error              { return nil }

//Personal.AI order the ending