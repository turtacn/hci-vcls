package fdm

import (
	"context"
	"testing"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
)

type mockAgent struct {
	id     string
	level  DegradationLevel
	started bool
	stopped bool
}

func (m *mockAgent) Start(ctx context.Context) error { m.started = true; return nil }
func (m *mockAgent) Stop() error                     { m.stopped = true; return nil }
func (m *mockAgent) LocalDegradationLevel() DegradationLevel { return m.level }
func (m *mockAgent) ClusterView() ClusterView        { return ClusterView{} }
func (m *mockAgent) IsLeader() bool                  { return false }
func (m *mockAgent) LeaderNodeID() string            { return "" }
func (m *mockAgent) NodeStates() map[string]NodeState { return nil }
func (m *mockAgent) OnNodeFailure(func(string))      {}
func (m *mockAgent) OnDegradationChanged(func(DegradationLevel)) {}

func TestDaemonImpl(t *testing.T) {
	cfg := DaemonConfig{
		Agents: []AgentConfig{
			{FDMConfig: FDMConfig{NodeID: "agent1"}},
			{FDMConfig: FDMConfig{NodeID: "agent2"}},
		},
	}
	daemon := NewDaemon(cfg, logger.Default(), metrics.NewNoopMetrics())

	a1 := &mockAgent{id: "agent1", level: DegradationMinor}
	a2 := &mockAgent{id: "agent2", level: DegradationMajor}

	daemon.(*daemonImpl).RegisterAgent("agent1", a1)
	daemon.(*daemonImpl).RegisterAgent("agent2", a2)

	// Test Agent retrieval
	if _, err := daemon.Agent("agent1"); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if _, err := daemon.Agent("nonexistent"); err == nil {
		t.Error("expected error for nonexistent agent")
	}

	// Test Agents map
	if len(daemon.Agents()) != 2 {
		t.Errorf("expected 2 agents, got %d", len(daemon.Agents()))
	}

	// Test ClusterDegradationLevel
	level := daemon.ClusterDegradationLevel()
	if string(level) == "" {
		t.Errorf("expected non-empty degradation string, got %v", level)
	}

	// Test Start
	err := daemon.Start(context.Background())
	if err != nil {
		t.Errorf("expected nil error on start, got %v", err)
	}
	if !a1.started || !a2.started {
		t.Error("expected agents to be started")
	}

	// Test Callbacks
	cbCalled := false
	daemon.OnAnyNodeFailure(func(nodeID, agentID string) { cbCalled = true })
	// We simulate a failure by invoking it
	if len(daemon.(*daemonImpl).cbs) == 1 {
		daemon.(*daemonImpl).cbs[0]("node1", "agent1")
		if !cbCalled {
			t.Error("expected cb to be called")
		}
	} else {
		t.Error("expected 1 callback registered")
	}

	// Test Stop
	err = daemon.Stop()
	if err != nil {
		t.Errorf("expected nil error on stop, got %v", err)
	}
	if !a1.stopped || !a2.stopped {
		t.Error("expected agents to be stopped")
	}
}
