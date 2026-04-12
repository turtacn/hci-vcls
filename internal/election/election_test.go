package election

import (
	"context"
	"testing"
	"time"
)

func TestMemoryElector(t *testing.T) {
	e := NewMemoryElector("node-1", nil)
	e.SetNodesCount(1) // Single node cluster becomes leader immediately

	if e.IsLeader() {
		t.Errorf("Expected initial state to not be leader")
	}

	watchCh := e.Watch()

	err := e.Campaign(context.Background())
	if err != nil {
		t.Fatalf("Campaign failed: %v", err)
	}

	if !e.IsLeader() {
		t.Errorf("Expected to be leader after campaign")
	}

	select {
	case status := <-watchCh:
		if !status.IsLeader || status.LeaderID != "node-1" {
			t.Errorf("Unexpected watch status: %+v", status)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timeout waiting for watch event")
	}

	err = e.Resign(context.Background())
	if err != nil {
		t.Fatalf("Resign failed: %v", err)
	}

	if e.IsLeader() {
		t.Errorf("Expected to not be leader after resign")
	}

	select {
	case status := <-watchCh:
		if status.IsLeader || status.LeaderID != "" {
			t.Errorf("Unexpected watch status after resign: %+v", status)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timeout waiting for watch event")
	}
}

func TestMemoryElector_Competition(t *testing.T) {
	e1 := NewMemoryElector("node-1", nil)
	e2 := NewMemoryElector("node-2", nil)
	e3 := NewMemoryElector("node-3", nil)

	e1.SetNodesCount(3)
	e2.SetNodesCount(3)
	e3.SetNodesCount(3)

	_ = e1.Campaign(context.Background())

	// node-1 campaigned, so it should have term=1, votedFor="node-1"
	t1, v1, l1 := e1.CurrentTermAndVote()
	if t1 != 1 || v1 != "node-1" || l1 {
		t.Errorf("e1 unexpected term/vote/leader: %d/%s/%v", t1, v1, l1)
	}

	// Send e1's campaign state (votedFor itself) to e2 and e3
	e2.ReceivePeerState("node-1", t1, "node-1", false)
	e3.ReceivePeerState("node-1", t1, "node-1", false)

	// e2 and e3 should have updated their term and voted for node-1
	t2, v2, _ := e2.CurrentTermAndVote()
	if t2 != 1 || v2 != "node-1" {
		t.Errorf("e2 unexpected term/vote after receiving: %d/%s", t2, v2)
	}

	// feed e2 and e3's votes back to e1
	e1.ReceivePeerState("node-2", t2, v2, false)
	t3, v3, _ := e3.CurrentTermAndVote()
	e1.ReceivePeerState("node-3", t3, v3, false)

	if !e1.IsLeader() {
		t.Errorf("Expected node-1 to win election after getting votes")
	}
	if e2.IsLeader() {
		t.Errorf("Expected node-2 to not be leader")
	}
}

func TestMemoryElector_RegressionT1(t *testing.T) {
	e1 := NewMemoryElector("node-1", nil)

	e1.SetNodesCount(3)

	// Become leader
	_ = e1.Campaign(context.Background())
	e1.ReceivePeerState("node-2", 1, "node-1", false)

	if !e1.IsLeader() {
		t.Errorf("Expected node-1 to be leader")
	}

	// Receive higher term
	e1.ReceivePeerState("node-2", 2, "node-2", false)

	if e1.IsLeader() {
		t.Errorf("Expected node-1 to drop leadership on higher term")
	}

	term, vote, _ := e1.CurrentTermAndVote()
	if term != 2 {
		t.Errorf("Expected term 2, got %d", term)
	}
	if vote != "node-2" {
		t.Errorf("Expected vote node-2, got %v", vote)
	}
}
