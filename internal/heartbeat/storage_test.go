package heartbeat

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStorageHeartbeater(t *testing.T) {
	dir := t.TempDir()

	cfg := HeartbeatConfig{
		NodeID:     "node1",
		Peers:      []string{"node2"},
		IntervalMs: 10,
		TimeoutMs:  50,
	}

	hb := NewStorageHeartbeater(cfg, dir)

	err := hb.Start(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Update digest
	hb.UpdateDigest(1, "node1")

	// Wait for write loop
	time.Sleep(30 * time.Millisecond)

	// Verify file was written
	filename := filepath.Join(dir, "node1.hb")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("expected hb file to be created")
	}

	// Simulate peer heartbeat file
	peerFile := filepath.Join(dir, "node2.hb")
	os.WriteFile(peerFile, []byte(`{"NodeID":"node2","Timestamp":"2099-01-01T00:00:00Z"}`), 0644)

	// Wait for read loop
	time.Sleep(30 * time.Millisecond)

	state, err := hb.PeerState("node2")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if !state.IsAlive {
		t.Error("expected peer to be alive")
	}

	err = hb.Stop()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestStorageHeartbeater_PeerStateError(t *testing.T) {
	hb := NewStorageHeartbeater(HeartbeatConfig{}, t.TempDir())
	_, err := hb.PeerState("node2")
	if err != ErrPeerNotFound {
		t.Errorf("expected ErrPeerNotFound, got %v", err)
	}
}
