package vcls

import (
	"context"
	"errors"
	"testing"

	"github.com/turtacn/hci-vcls/pkg/fdm"
)

type mockFDMAgent struct {
	level fdm.DegradationLevel
	cb    func(fdm.DegradationLevel)
}

func (m *mockFDMAgent) Start(ctx context.Context) error                         { return nil }
func (m *mockFDMAgent) Stop() error                                             { return nil }
func (m *mockFDMAgent) NodeStates() map[string]fdm.NodeState                    { return nil }
func (m *mockFDMAgent) LocalDegradationLevel() fdm.DegradationLevel             { return m.level }
func (m *mockFDMAgent) IsLeader() bool                                          { return false }
func (m *mockFDMAgent) LeaderNodeID() string                                    { return "" }
func (m *mockFDMAgent) OnNodeFailure(callback func(nodeID string))              {}
func (m *mockFDMAgent) OnDegradationChanged(callback func(level fdm.DegradationLevel)) {
	m.cb = callback
}
func (m *mockFDMAgent) ClusterView() fdm.ClusterView { return fdm.ClusterView{} }

func TestAgent_StartStop(t *testing.T) {
	config := VCLSConfig{Matrix: DefaultCapabilityMatrix()}
	agent := NewAgent(config, &mockFDMAgent{level: fdm.DegradationNone})

	ctx := context.Background()

	err := agent.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed")
	}

	state := agent.ClusterServiceState()
	if state != ServiceStateHealthy {
		t.Errorf("Expected healthy state, got %v", state)
	}

	caps := agent.ActiveCapabilities()
	if len(caps) != 6 {
		t.Errorf("Expected 6 active capabilities for DegradationNone, got %d", len(caps))
	}

	err = agent.Stop()
	if err != nil {
		t.Fatalf("Expected stop to succeed")
	}
}

func TestAgent_Degradation(t *testing.T) {
	config := VCLSConfig{Matrix: DefaultCapabilityMatrix()}
	fdmAgent := &mockFDMAgent{level: fdm.DegradationNone}
	agent := NewAgent(config, fdmAgent)

	ctx := context.Background()

	_ = agent.Start(ctx)

	// Simulate degradation to ZK
	fdmAgent.cb(fdm.DegradationZK)

	state := agent.ClusterServiceState()
	if state != ServiceStateDegraded {
		t.Errorf("Expected degraded state, got %v", state)
	}

	caps := agent.ActiveCapabilities()
	if len(caps) != 3 {
		t.Errorf("Expected 3 active capabilities for DegradationZK, got %d", len(caps))
	}

	if !agent.IsCapable(CapabilityHA) {
		t.Errorf("Expected HA capability to be active")
	}
	if agent.IsCapable(CapabilityFT) {
		t.Errorf("Expected FT capability to be inactive")
	}

	// Simulate degradation to All
	fdmAgent.cb(fdm.DegradationAll)

	state = agent.ClusterServiceState()
	if state != ServiceStateOffline {
		t.Errorf("Expected offline state, got %v", state)
	}

	if len(agent.ActiveCapabilities()) != 0 {
		t.Errorf("Expected 0 active capabilities for DegradationAll")
	}

	err := agent.RequireCapability(CapabilityHA)
	if err != ErrCapabilityUnavailable {
		t.Errorf("Expected ErrCapabilityUnavailable, got %v", err)
	}
}

func TestAgent_InvalidMatrix(t *testing.T) {
	config := VCLSConfig{Matrix: CapabilityMatrix{}} // Missing DegradationNone
	agent := NewAgent(config, &mockFDMAgent{})

	ctx := context.Background()

	err := agent.Start(ctx)
	if err != ErrInvalidCapabilityMatrix {
		t.Errorf("Expected ErrInvalidCapabilityMatrix, got %v", err)
	}
}

func TestCapabilityError(t *testing.T) {
	err := ErrCapabilityUnavailable
	if err.Error() == "" {
		t.Errorf("Expected error message to not be empty")
	}

	wrappedErr := &CapabilityError{Code: "ERR_CUSTOM", Message: "custom error", Err: errors.New("inner error")}
	if err := wrappedErr.Unwrap(); err == nil {
		t.Errorf("Expected to unwrap an inner error")
	}
}

//Personal.AI order the ending