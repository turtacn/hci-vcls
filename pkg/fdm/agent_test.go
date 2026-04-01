package fdm

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
)

type mockProber struct {
	results map[HeartbeatLevel]ProbeResult
}

func (m *mockProber) ProbeL0(ctx context.Context) ProbeResult { return m.results[HeartbeatL0] }
func (m *mockProber) ProbeL1(ctx context.Context) ProbeResult { return m.results[HeartbeatL1] }
func (m *mockProber) ProbeL2(ctx context.Context) ProbeResult { return m.results[HeartbeatL2] }
func (m *mockProber) ProbeAll(ctx context.Context) map[HeartbeatLevel]ProbeResult {
	return m.results
}

type mockElector struct {
	isLeader  bool
	leaderID  string
	callbacks []func(election.LeaderInfo)
}

func (m *mockElector) Campaign(ctx context.Context) error {
	m.isLeader = true
	m.leaderID = "node-1"
	for _, cb := range m.callbacks {
		cb(election.LeaderInfo{NodeID: m.leaderID, Term: 1})
	}
	return nil
}

func (m *mockElector) Resign(ctx context.Context) error { return nil }
func (m *mockElector) IsLeader() bool                   { return m.isLeader }
func (m *mockElector) LeaderID() (string, error) {
	if m.leaderID == "" {
		return "", errors.New("no leader")
	}
	return m.leaderID, nil
}
func (m *mockElector) OnLeaderChange(callback func(election.LeaderInfo)) {
	m.callbacks = append(m.callbacks, callback)
}
func (m *mockElector) Close() error { return nil }

func TestAgent_StartStop_Election(t *testing.T) {
	config := FDMConfig{
		NodeID:          "node-1",
		ClusterID:       "cluster-1",
		HeartbeatPeers:  []string{"node-2"},
		ProbeIntervalMs: 100,
	}

	prober := &mockProber{
		results: map[HeartbeatLevel]ProbeResult{
			HeartbeatL0: {Level: HeartbeatL0, Success: true},
		},
	}
	elector := &mockElector{}
	log := logger.Default()
	m := metrics.NewNoopMetrics()

	agent := NewAgent(config, prober, elector, log, m)
	ctx := context.Background()

	err := agent.Start(ctx)
	if err != nil {
		t.Fatalf("Expected start to succeed, got %v", err)
	}

	if !agent.IsLeader() {
		t.Errorf("Expected agent to be leader after campaign")
	}

	if agent.LeaderNodeID() != "node-1" {
		t.Errorf("Expected leader ID to be node-1, got %s", agent.LeaderNodeID())
	}

	err = agent.Stop()
	if err != nil {
		t.Fatalf("Expected stop to succeed, got %v", err)
	}
}

func TestAgent_Degradation(t *testing.T) {
	config := FDMConfig{
		NodeID:          "node-1",
		ProbeIntervalMs: 10,
	}

	prober := &mockProber{
		results: map[HeartbeatLevel]ProbeResult{
			HeartbeatL0: {Level: HeartbeatL0, Success: false},
			HeartbeatL1: {Level: HeartbeatL1, Success: false},
			HeartbeatL2: {Level: HeartbeatL2, Success: false},
		},
	}
	elector := &mockElector{}
	log := logger.Default()
	m := metrics.NewNoopMetrics()

	agent := NewAgent(config, prober, elector, log, m)
	ctx := context.Background()

	degChan := make(chan DegradationLevel, 1)
	agent.OnDegradationChanged(func(level DegradationLevel) {
		degChan <- level
	})

	_ = agent.Start(ctx)
	defer func() { _ = agent.Stop() }()

	// Wait for degradation due to all probes failing
	select {
	case level := <-degChan:
		if level != DegradationAll {
			t.Errorf("Expected DegradationAll, got %v", level)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Timeout waiting for degradation callback")
	}

	if agent.LocalDegradationLevel() != DegradationAll {
		t.Errorf("Expected agent level to be DegradationAll")
	}

	// Change prober to succeed
	prober.results = map[HeartbeatLevel]ProbeResult{
		HeartbeatL0: {Level: HeartbeatL0, Success: true},
	}

	// Wait for recovery
	select {
	case level := <-degChan:
		if level != DegradationNone {
			t.Errorf("Expected DegradationNone, got %v", level)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Timeout waiting for recovery callback")
	}
}

func TestFDMError(t *testing.T) {
	err := ErrNodeNotFound
	if err.Error() == "" {
		t.Errorf("Expected error message to not be empty")
	}

	wrappedErr := &FDMError{Code: "ERR_CUSTOM", Message: "custom error", Err: errors.New("inner error")}
	if err := wrappedErr.Unwrap(); err == nil {
		t.Errorf("Expected to unwrap an inner error")
	}
}

//Personal.AI order the ending