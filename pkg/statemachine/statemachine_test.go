package statemachine

import (
	"testing"
)

func TestMachine(t *testing.T) {
	m := NewMachine()

	if m.Current() != StateInit {
		t.Errorf("Expected initial state StateInit, got %v", m.Current())
	}

	// Test invalid transition
	if m.CanTransition(EventFailoverCompleted) {
		t.Errorf("Should not be able to transition EventFailoverCompleted from StateInit")
	}
	err := m.Transition(EventFailoverCompleted)
	if err != ErrIllegalTransition {
		t.Errorf("Expected ErrIllegalTransition, got %v", err)
	}

	// Test valid transitions
	events := []Event{
		EventHeartbeatRestored,   // Init -> Stable
		EventDegradationDetected, // Stable -> Degraded
		EventEvaluationStarted,   // Degraded -> Evaluating
		EventFailoverTriggered,   // Evaluating -> Failover
		EventFailoverCompleted,   // Failover -> Recovered
		EventHeartbeatRestored,   // Recovered -> Stable
	}

	for _, e := range events {
		if !m.CanTransition(e) {
			t.Errorf("Expected CanTransition %v from %v to be true", e, m.Current())
		}
		err := m.Transition(e)
		if err != nil {
			t.Errorf("Failed to transition %v: %v", e, err)
		}
	}

	if m.Current() != StateStable {
		t.Errorf("Expected final state StateStable, got %v", m.Current())
	}

	history := m.History()
	if len(history) != len(events) {
		t.Errorf("Expected history length %d, got %d", len(events), len(history))
	}
}

// Personal.AI order the ending
