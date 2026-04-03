package election

import (
	"context"
	"testing"
	"time"
)

func TestMemoryElector(t *testing.T) {
	// reset global state for test
	globalLeaderLock.Lock()
	currentLeader = ""
	globalLeaderLock.Unlock()

	e := NewMemoryElector("node-1")

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
	globalLeaderLock.Lock()
	currentLeader = ""
	globalLeaderLock.Unlock()

	e1 := NewMemoryElector("node-1")
	e2 := NewMemoryElector("node-2")

	_ = e1.Campaign(context.Background())
	_ = e2.Campaign(context.Background())

	if !e1.IsLeader() {
		t.Errorf("Expected node-1 to win election")
	}
	if e2.IsLeader() {
		t.Errorf("Expected node-2 to lose election")
	}
}

// Personal.AI order the ending
