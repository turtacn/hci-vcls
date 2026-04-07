package heartbeat

import (
	"context"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/statemachine"
)

type mockElector struct {
	leader bool
}

func (m *mockElector) Campaign(ctx context.Context) error { return nil }
func (m *mockElector) Resign(ctx context.Context) error   { return nil }
func (m *mockElector) Close() error                       { return nil }
func (m *mockElector) OnLeaderChange(cb func(election.LeaderInfo)) {}
func (m *mockElector) IsLeader() bool { return m.leader }
func (m *mockElector) Status() election.LeaderStatus { return election.LeaderStatus{LeaderID: "node1"} }
func (m *mockElector) Watch() <-chan election.LeaderStatus { return nil }

type mockEvaluator struct {
	state *fdm.ClusterState
}

func (m *mockEvaluator) Evaluate(ctx context.Context, clusterID string, leaderNodeID string, hosts []fdm.HostState, witnessAvailable bool) (*fdm.ClusterState, error) {
	return m.state, nil
}

func TestHeartbeatService(t *testing.T) {
	cfg := HeartbeatConfig{IntervalMs: 10, TimeoutMs: 20}
	monitor := NewMemoryMonitor()
	elector := &mockElector{leader: true}
	sm := statemachine.NewMachine()

	evaluator := &mockEvaluator{
		state: &fdm.ClusterState{Degradation: fdm.DegradationMinor},
	}

	svc := NewService(cfg, monitor, elector, evaluator, sm, metrics.NewNoopMetrics(), logger.Default())

	// Add a summary
	monitor.Record(Sample{NodeID: "node2", ClusterID: "cluster-1", ReceivedAt: time.Now()})

	err := svc.Start()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	time.Sleep(50 * time.Millisecond) // Let loop run and timeout happen

	// Since we haven't updated the summary, it should timeout
	if sm.Current() != statemachine.StateDegraded {
		t.Errorf("expected Degraded state, got %v", sm.Current())
	}

	err = svc.Stop()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestHeartbeatService_NotLeader(t *testing.T) {
	cfg := HeartbeatConfig{IntervalMs: 10, TimeoutMs: 20}
	monitor := NewMemoryMonitor()
	elector := &mockElector{leader: false}

	// Evaluator should not be called
	evaluator := &mockEvaluator{state: nil}

	svc := NewService(cfg, monitor, elector, evaluator, nil, nil, logger.Default())

	monitor.Record(Sample{NodeID: "node2", ClusterID: "cluster-1", ReceivedAt: time.Now()})

	_ = svc.Start()
	time.Sleep(50 * time.Millisecond)
	_ = svc.Stop()
}
