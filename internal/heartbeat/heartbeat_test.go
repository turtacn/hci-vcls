package heartbeat

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/pkg/witness"
)

type mockWitnessPool struct{}

func (m mockWitnessPool) ConfirmFailure(ctx context.Context, req witness.ConfirmationRequest) bool {
	return true
}

func (m mockWitnessPool) Quorum(ctx context.Context) bool {
	return true
}

func (m mockWitnessPool) Statuses(ctx context.Context) map[string]witness.WitnessStatus {
	return nil
}

func TestHeartbeater_StartStop(t *testing.T) {
	config := HeartbeatConfig{
		IntervalMs: 10,
		TimeoutMs:  50,
		NodeID:     "node-1",
		Peers:      []string{"node-2"},
	}

	hb := NewHeartbeater(config, mockWitnessPool{})
	ctx := context.Background()

	err := hb.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start heartbeater: %v", err)
	}

	state, err := hb.PeerState("node-2")
	if err != nil {
		t.Fatalf("Failed to get peer state: %v", err)
	}
	if !state.IsAlive {
		t.Errorf("Expected peer to be alive initially")
	}

	err = hb.Stop()
	if err != nil {
		t.Fatalf("Failed to stop heartbeater: %v", err)
	}

	states := hb.AllPeerStates()
	if len(states) != 1 {
		t.Errorf("Expected 1 peer state, got %d", len(states))
	}
}

func TestHeartbeater_PeerTimeout(t *testing.T) {
	config := HeartbeatConfig{
		IntervalMs: 10,
		TimeoutMs:  20, // Short timeout for test
		NodeID:     "node-1",
		Peers:      []string{"node-2"},
	}

	hb := NewHeartbeater(config, mockWitnessPool{})
	ctx := context.Background()

	deadChan := make(chan string, 1)
	hb.OnPeerDead(func(nodeID string) {
		deadChan <- nodeID
	})

	err := hb.Start(ctx)
	if err != nil {
		t.Fatalf("Failed to start heartbeater: %v", err)
	}
	defer func() { _ = hb.Stop() }()

	// Wait for peer to timeout (TimeoutMs is 20ms, IntervalMs is 10ms)
	select {
	case nodeID := <-deadChan:
		if nodeID != "node-2" {
			t.Errorf("Expected dead callback for node-2, got %s", nodeID)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("Timeout waiting for peer death callback")
	}

	state, err := hb.PeerState("node-2")
	if err != nil {
		t.Fatalf("Failed to get peer state: %v", err)
	}
	if state.IsAlive {
		t.Errorf("Expected peer to be marked as dead")
	}
}

func TestHeartbeatError(t *testing.T) {
	err := ErrPeerNotFound
	if err.Error() == "" {
		t.Errorf("Expected error message to not be empty")
	}

	wrappedErr := &HeartbeatError{Code: "ERR_CUSTOM", Message: "custom error", Err: errors.New("inner error")}
	if err := wrappedErr.Unwrap(); err == nil {
		t.Errorf("Expected to unwrap an inner error")
	}
}

