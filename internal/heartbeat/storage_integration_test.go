package heartbeat

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
)

func TestStorageHeartbeater_Integration(t *testing.T) {
	dir := t.TempDir()
	log := logger.Default()

	cfg1 := HeartbeatConfig{NodeID: "node1", Peers: []string{"node2"}, IntervalMs: 10}
	hb1 := NewStorageHeartbeater(cfg1, dir, log)

	cfg2 := HeartbeatConfig{NodeID: "node2", Peers: []string{"node1"}, IntervalMs: 10}
	hb2 := NewStorageHeartbeater(cfg2, dir, log)

	// Test Write & Read for node1
	now := time.Now().Round(time.Second) // Rounding to avoid precision issues in JSON
	hb1.Write("node1", now)

	ts, err := hb2.Read("node1")
	if err != nil {
		t.Fatalf("expected no error reading node1, got %v", err)
	}
	if !ts.Equal(now) {
		t.Errorf("expected read timestamp %v to equal %v", ts, now)
	}

	// Test Write & Read for node2
	later := now.Add(time.Second)
	hb2.Write("node2", later)

	ts, err = hb1.Read("node2")
	if err != nil {
		t.Fatalf("expected no error reading node2, got %v", err)
	}
	if !ts.Equal(later) {
		t.Errorf("expected read timestamp %v to equal %v", ts, later)
	}

	// Test ReadAll
	allTimes := hb1.ReadAll()
	if len(allTimes) != 2 {
		t.Errorf("expected 2 records from ReadAll, got %d", len(allTimes))
	}
	if !allTimes["node1"].Equal(now) || !allTimes["node2"].Equal(later) {
		t.Errorf("readall results mismatch, got: %v", allTimes)
	}

	// Test atomic failure resilience (corrupted file)
	node3Dir := filepath.Join(dir, "node3")
	os.MkdirAll(node3Dir, 0755)
	tmpPath := filepath.Join(node3Dir, "hb.json")
	os.WriteFile(tmpPath, []byte("invalid-json"), 0644)

	_, err = hb1.Read("node3")
	if err == nil {
		t.Error("expected error reading invalid json file, got nil")
	}

	// Ensure ReadAll safely ignores corrupted file
	allTimes = hb1.ReadAll()
	if len(allTimes) != 2 {
		t.Errorf("expected ReadAll to ignore corrupted node3, got %d records", len(allTimes))
	}

	// Test Monitor complementing L1 UDP via StorageHeartbeat
	monitor := NewMemoryMonitor(hb1)
	monitor.Record(Sample{NodeID: "node2", ClusterID: "c1", ReceivedAt: now.Add(-5 * time.Second)})

	// Wait, if node2 sent L1 UDP 5 seconds ago, normally it would be timed out
	// But `node2` has a fresh timestamp in L2 storage (`later`).
	// We'll call CheckTimeouts with a timeout of 2 seconds.
	monitor.CheckTimeouts(later.Add(time.Second), 2*time.Second)

	summary, exists := monitor.GetSummary("node2")
	if !exists {
		t.Fatal("expected summary for node2 to exist")
	}

	// Because of L2 storage `later` timestamp, it should be marked fresh again and healthy
	if !summary.Healthy {
		t.Error("expected node2 to remain healthy due to L2 storage heartbeat")
	}
	if !summary.LastSeenAt.Equal(later) {
		t.Errorf("expected LastSeenAt to be updated from L2 storage (%v), got %v", later, summary.LastSeenAt)
	}

	// Wait more so even L2 is timed out
	monitor.CheckTimeouts(later.Add(10*time.Second), 2*time.Second)
	summary, _ = monitor.GetSummary("node2")
	if summary.Healthy {
		t.Error("expected node2 to be unhealthy after both L1 and L2 timed out")
	}
}
