package heartbeat

import (
	"testing"
	"time"
)

func TestMemoryMonitor(t *testing.T) {
	monitor := NewMemoryMonitor()

	now := time.Now()
	sample := Sample{
		NodeID: "node1",
		ClusterID: "cluster1",
		ReceivedAt: now,
	}

	monitor.Record(sample)

	summary, exists := monitor.GetSummary("node1")
	if !exists {
		t.Fatal("expected summary to exist")
	}
	if summary.NodeID != "node1" {
		t.Errorf("expected node1, got %v", summary.NodeID)
	}
	if !summary.Healthy {
		t.Error("expected summary to be healthy")
	}

	_, exists = monitor.GetSummary("node2")
	if exists {
		t.Error("expected summary not to exist")
	}

	summaries := monitor.ListSummaries("cluster1")
	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
	}

	summaries = monitor.ListSummaries("cluster2")
	if len(summaries) != 0 {
		t.Errorf("expected 0 summaries, got %d", len(summaries))
	}

	// Test timeout
	monitor.CheckTimeouts(now.Add(2*time.Second), 1*time.Second)
	summary, _ = monitor.GetSummary("node1")
	if summary.Healthy {
		t.Error("expected summary to be unhealthy after timeout")
	}
	if summary.LostCount != 1 {
		t.Errorf("expected lost count 1, got %d", summary.LostCount)
	}

	// Double timeout shouldn't increment lost count immediately if already false, wait, CheckTimeouts logic:
	// `if s.Healthy && now.Sub(s.LastSeenAt) > timeout { s.Healthy = false; s.LostCount++ }`
	monitor.CheckTimeouts(now.Add(3*time.Second), 1*time.Second)
	summary, _ = monitor.GetSummary("node1")
	if summary.LostCount != 1 {
		t.Errorf("expected lost count 1, got %d", summary.LostCount)
	}
}
