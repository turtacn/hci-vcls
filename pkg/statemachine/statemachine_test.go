package statemachine

import (
	"errors"
	"testing"
)

func TestStateMachine_ValidTransitions(t *testing.T) {
	sm := NewStateMachine()

	if sm.Current() != StateInit {
		t.Fatalf("expected initial state %s, got %s", StateInit, sm.Current())
	}

	err := sm.Event(EventHeartbeatRestored)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.Current() != StateStable {
		t.Fatalf("expected %s, got %s", StateStable, sm.Current())
	}

	err = sm.Event(EventDegradationDetected)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.Current() != StateDegraded {
		t.Fatalf("expected %s, got %s", StateDegraded, sm.Current())
	}

	err = sm.Event(EventEvaluationStarted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.Current() != StateEvaluating {
		t.Fatalf("expected %s, got %s", StateEvaluating, sm.Current())
	}

	err = sm.Event(EventFailoverTriggered)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.Current() != StateFailover {
		t.Fatalf("expected %s, got %s", StateFailover, sm.Current())
	}

	err = sm.Event(EventFailoverCompleted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.Current() != StateRecovered {
		t.Fatalf("expected %s, got %s", StateRecovered, sm.Current())
	}

	err = sm.Event(EventHeartbeatRestored)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sm.Current() != StateStable {
		t.Fatalf("expected %s, got %s", StateStable, sm.Current())
	}
}

func TestStateMachine_InvalidTransition(t *testing.T) {
	sm := NewStateMachine()

	err := sm.Event(EventFailoverCompleted)
	if err == nil {
		t.Fatalf("expected error for invalid transition, got nil")
	}

	if !errors.Is(err, ErrInvalidTransition) {
		t.Fatalf("expected %v, got %v", ErrInvalidTransition, err)
	}

	if sm.Current() != StateInit {
		t.Fatalf("expected state to remain %s, got %s", StateInit, sm.Current())
	}
}

//Personal.AI order the ending
