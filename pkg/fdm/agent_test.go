package fdm

import (
	"context"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/logger"
)

import "sync"

type mockProber struct {
	mu      sync.Mutex
	results map[HeartbeatLevel]ProbeResult
}

func (m *mockProber) ProbeL0(ctx context.Context) ProbeResult { return ProbeResult{} }
func (m *mockProber) ProbeL1(ctx context.Context) ProbeResult { return ProbeResult{} }
func (m *mockProber) ProbeL2(ctx context.Context) ProbeResult { return ProbeResult{} }

func (m *mockProber) ProbeAll(ctx context.Context) map[HeartbeatLevel]ProbeResult {
	m.mu.Lock()
	defer m.mu.Unlock()
	res := make(map[HeartbeatLevel]ProbeResult)
	for k, v := range m.results {
		res[k] = v
	}
	return res
}

type mockElector struct {
	mu sync.RWMutex
	cb func(election.LeaderInfo)
}

func (m *mockElector) Campaign(ctx context.Context) error { return nil }
func (m *mockElector) Resign(ctx context.Context) error   { return nil }
func (m *mockElector) Close() error                       { return nil }
func (m *mockElector) OnLeaderChange(cb func(election.LeaderInfo)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cb = cb
}
func (m *mockElector) ReceivePeerState(peerNodeID string, peerTerm int64, peerVoteFor string, isLeader bool) {}
func (m *mockElector) CurrentTermAndVote() (int64, string, bool) { return 0, "", false }
func (m *mockElector) SetNodesCount(count int) {}
func (m *mockElector) TriggerLeaderChange(info election.LeaderInfo) {
	m.mu.RLock()
	cb := m.cb
	m.mu.RUnlock()
	if cb != nil {
		cb(info)
	}
}
func (m *mockElector) IsLeader() bool                      { return false }
func (m *mockElector) LeaderID() string                    { return "" }
func (m *mockElector) Status() election.LeaderStatus       { return election.LeaderStatus{} }
func (m *mockElector) Watch() <-chan election.LeaderStatus { return nil }

type mockMetrics struct{}

func (m *mockMetrics) IncElectionTotal(node, result string)                       {}
func (m *mockMetrics) IncLeaderChange(cluster string)                             {}
func (m *mockMetrics) IncHeartbeatLost(node, cluster string)                      {}
func (m *mockMetrics) SetDegradationLevel(cluster string, level float64)          {}
func (m *mockMetrics) IncHATaskTotal(cluster, status string)                      {}
func (m *mockMetrics) ObserveHAExecutionDuration(cluster string, seconds float64) {}
func (m *mockMetrics) SetProtectedVMCount(cluster string, count float64)          {}

func TestAgent(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	m := &mockMetrics{}
	elector := &mockElector{}

	prober := &mockProber{
		results: map[HeartbeatLevel]ProbeResult{
			HeartbeatL0: {Success: true},
		},
	}

	config := FDMConfig{
		ClusterID:       "cluster-1",
		NodeID:          "node-1",
		HeartbeatPeers:  []string{"peer-1"},
		ProbeIntervalMs: 10,
	}

	agent := NewAgent(config, prober, elector, log, m)

	if agent.IsLeader() {
		t.Errorf("Should be false initially")
	}

	states := agent.NodeStates()
	if states["peer-1"] != NodeStateAlive {
		t.Errorf("Expected peer-1 to be alive, got %v", states["peer-1"])
	}

	ctx := context.Background()
	err := agent.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start agent: %v", err)
	}

	// Trigger leader change
	elector.TriggerLeaderChange(election.LeaderInfo{NodeID: "node-1"})
	if !agent.IsLeader() {
		t.Errorf("Expected to be leader")
	}
	if agent.LeaderNodeID() != "node-1" {
		t.Errorf("Expected leader to be node-1")
	}

	agent.OnNodeFailure(func(nodeID string) {})
	agent.OnDegradationChanged(func(level DegradationLevel) {})

	time.Sleep(50 * time.Millisecond) // Let probe loop run

	view := agent.ClusterView()
	if view.DegradationLevel != OldDegradationNone {
		t.Errorf("Expected OldDegradationNone, got %v", view.DegradationLevel)
	}

	// Change prober to return failures
	prober.mu.Lock()
	prober.results = map[HeartbeatLevel]ProbeResult{
		HeartbeatL0: {Success: false},
	}
	prober.mu.Unlock()

	time.Sleep(50 * time.Millisecond) // Let probe loop run

	level := agent.LocalDegradationLevel()
	if level != DegradationCritical {
		t.Errorf("Expected DegradationCritical, got %v", level)
	}

	err = agent.Stop()
	if err != nil {
		t.Fatalf("Failed to stop agent: %v", err)
	}
}

