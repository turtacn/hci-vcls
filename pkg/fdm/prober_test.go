package fdm

import (
	"context"
	"testing"

	"github.com/turtacn/hci-vcls/internal/heartbeat"
	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/witness"
)

type mockHeartbeater struct{}

func (m *mockHeartbeater) Start(ctx context.Context) error { return nil }
func (m *mockHeartbeater) Stop() error                     { return nil }
func (m *mockHeartbeater) PeerState(nodeID string) (heartbeat.HeartbeatState, error) {
	return heartbeat.HeartbeatState{}, nil
}
func (m *mockHeartbeater) AllPeerStates() map[string]heartbeat.HeartbeatState { return nil }
func (m *mockHeartbeater) OnPeerDead(callback func(nodeID string))            {}
func (m *mockHeartbeater) OnPeerRecovered(callback func(nodeID string))       {}

type mockWitnessPool struct{}

func (m *mockWitnessPool) ConfirmFailure(ctx context.Context, req witness.ConfirmationRequest) bool {
	return true
}
func (m *mockWitnessPool) Quorum(ctx context.Context) bool { return true }
func (m *mockWitnessPool) Statuses(ctx context.Context) map[string]witness.WitnessStatus {
	return nil
}

func TestProber(t *testing.T) {
	hb := &mockHeartbeater{}
	pool := &mockWitnessPool{}
	log := logger.Default()
	m := metrics.NewNoopMetrics()

	prober := NewProber(hb, pool, log, m)
	ctx := context.Background()

	resL0 := prober.ProbeL0(ctx)
	if !resL0.Success || resL0.Level != HeartbeatL0 {
		t.Errorf("Expected ProbeL0 to succeed, got %v", resL0)
	}

	resL1 := prober.ProbeL1(ctx)
	if !resL1.Success || resL1.Level != HeartbeatL1 {
		t.Errorf("Expected ProbeL1 to succeed, got %v", resL1)
	}

	resL2 := prober.ProbeL2(ctx)
	if !resL2.Success || resL2.Level != HeartbeatL2 {
		t.Errorf("Expected ProbeL2 to succeed, got %v", resL2)
	}

	allRes := prober.ProbeAll(ctx)
	if len(allRes) != 3 {
		t.Errorf("Expected 3 probe results, got %d", len(allRes))
	}
	if !allRes[HeartbeatL0].Success || !allRes[HeartbeatL1].Success || !allRes[HeartbeatL2].Success {
		t.Errorf("Expected all probes to succeed")
	}
}

//Personal.AI order the ending
