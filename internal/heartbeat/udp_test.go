package heartbeat

import (
	"context"
	"net"
	"testing"
	"time"
)

func TestUDPHeartbeater(t *testing.T) {
	// Find a free port
	addr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	conn, _ := net.ListenUDP("udp", addr)
	addr = conn.LocalAddr().(*net.UDPAddr)
	conn.Close() // Free it so heartbeater can use it

	cfg := HeartbeatConfig{
		NodeID:     addr.String(),
		Peers:      []string{"127.0.0.1:9999"},
		IntervalMs: 10,
		TimeoutMs:  50,
	}

	hb := NewUDPHeartbeater(cfg)

	err := hb.Start(context.Background())
	if err != nil {
		t.Fatalf("expected nil error on start, got %v", err)
	}

	hb.UpdateDigest(1, "node1")

	// wait for send loop
	time.Sleep(20 * time.Millisecond)

	state, err := hb.PeerState("127.0.0.1:9999")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if !state.IsAlive {
		t.Error("expected peer to be alive initially")
	}

	time.Sleep(60 * time.Millisecond) // Let it timeout

	state, _ = hb.PeerState("127.0.0.1:9999")
	if state.IsAlive {
		t.Error("expected peer to be dead after timeout")
	}

	err = hb.Stop()
	if err != nil {
		t.Fatalf("expected nil error on stop, got %v", err)
	}
}

func TestUDPHeartbeater_PeerStateError(t *testing.T) {
	hb := NewUDPHeartbeater(HeartbeatConfig{})
	_, err := hb.PeerState("node2")
	if err != ErrPeerNotFound {
		t.Errorf("expected ErrPeerNotFound, got %v", err)
	}
}
