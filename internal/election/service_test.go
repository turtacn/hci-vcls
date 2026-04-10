package election

import (
	"context"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
)

type mockElector struct {
	campaignCalled bool
	resignCalled   bool
	closeCalled    bool
	ch             chan LeaderStatus
}

func (m *mockElector) Campaign(ctx context.Context) error { m.campaignCalled = true; return nil }
func (m *mockElector) Resign(ctx context.Context) error   { m.resignCalled = true; return nil }
func (m *mockElector) Close() error                       { m.closeCalled = true; return nil }
func (m *mockElector) OnLeaderChange(cb func(LeaderInfo)) {}
func (m *mockElector) ReceivePeerState(peerNodeID string, peerTerm int64, peerVoteFor string, isLeader bool) {}
func (m *mockElector) CurrentTermAndVote() (int64, string, bool) { return 0, "", false }
func (m *mockElector) SetNodesCount(count int) {}
func (m *mockElector) IsLeader() bool                     { return false }
func (m *mockElector) Status() LeaderStatus               { return LeaderStatus{} }
func (m *mockElector) Watch() <-chan LeaderStatus {
	m.ch = make(chan LeaderStatus, 1)
	return m.ch
}

func TestElectionService(t *testing.T) {
	elector := &mockElector{}
	svc := NewService(elector, metrics.NewNoopMetrics(), logger.Default())

	err := svc.Start()
	if err != nil {
		t.Fatalf("expected nil error on start, got %v", err)
	}

	if !elector.campaignCalled {
		t.Error("expected campaign to be called")
	}

	// Trigger watch loop
	time.Sleep(10 * time.Millisecond) // Let watchLoop initialize channel
	if elector.ch != nil {
		elector.ch <- LeaderStatus{IsLeader: true, LeaderID: "node1"}
	}

	time.Sleep(10 * time.Millisecond)

	err = svc.Stop()
	if err != nil {
		t.Fatalf("expected nil error on stop, got %v", err)
	}

	if !elector.resignCalled {
		t.Error("expected resign to be called")
	}
	if !elector.closeCalled {
		t.Error("expected close to be called")
	}
}
