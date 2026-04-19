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

	// Create an explicit teardown function instead of relying purely on TempDir
	// This helps avoid Windows-like directory not empty errors if the background
	// processes are still running during cleanup.
	defer os.RemoveAll(dir)

	cfg := HeartbeatConfig{
		NodeID:     "node1",
		Peers:      []string{"node2"},
		IntervalMs: 10,
		TimeoutMs:  50,
	}

	hb := NewStorageHeartbeater(cfg, dir, nil)

	err := hb.Start(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	// Update digest
	hb.UpdateDigest(1, "node1", true)

	// Wait for write loop
	time.Sleep(30 * time.Millisecond)

	// Verify file was written
	filename := filepath.Join(dir, "node1", "hb.json")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Errorf("expected hb file to be created")
	}

	// Simulate peer heartbeat file
	peerDir := filepath.Join(dir, "node2")
	os.MkdirAll(peerDir, 0755)
	peerFile := filepath.Join(peerDir, "hb.json")
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

	// Ensure goroutines fully exit before TempDir cleanup is triggered
	time.Sleep(50 * time.Millisecond)
}

func TestStorageHeartbeater_PeerStateError(t *testing.T) {
	hb := NewStorageHeartbeater(HeartbeatConfig{}, t.TempDir(), nil)
	_, err := hb.PeerState("node2")
	if err != ErrPeerNotFound {
		t.Errorf("expected ErrPeerNotFound, got %v", err)
	}
}
