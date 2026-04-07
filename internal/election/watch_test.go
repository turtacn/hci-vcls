package election

import (
	"testing"
)

func TestWatcher(t *testing.T) {
	w := NewWatcher()

	ch1 := w.Subscribe()
	ch2 := w.Subscribe()

	w.Notify(LeaderStatus{IsLeader: true, LeaderID: "node1"})

	status := <-ch1
	if !status.IsLeader {
		t.Error("expected true")
	}

	status2 := <-ch2
	if status2.LeaderID != "node1" {
		t.Errorf("expected node1, got %v", status2.LeaderID)
	}

	w.Close()

	// Should be closed
	_, ok := <-ch1
	if ok {
		t.Error("expected channel to be closed")
	}

	// Double close should not panic
	w.Close()

	// Notify after close should do nothing
	w.Notify(LeaderStatus{})
}
