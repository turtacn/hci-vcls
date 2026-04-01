package statemachine

import (
	"fmt"
	"sync"
)

type stateMachineImpl struct {
	mu      sync.Mutex
	current State
}

func NewStateMachine() StateTransitionMachine {
	return &stateMachineImpl{
		current: StateInit,
	}
}

func (s *stateMachineImpl) Current() State {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.current
}

func (s *stateMachineImpl) Event(e Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next, ok := validStateTransitions[s.current][e]
	if !ok {
		return fmt.Errorf("%w: cannot process %s in state %s", ErrInvalidTransition, e, s.current)
	}

	s.current = next
	return nil
}

var validStateTransitions = map[State]map[Event]State{
	StateInit: {
		EventHeartbeatRestored: StateStable,
	},
	StateStable: {
		EventDegradationDetected: StateDegraded,
	},
	StateDegraded: {
		EventEvaluationStarted: StateEvaluating,
		EventHeartbeatRestored: StateStable,
	},
	StateEvaluating: {
		EventFailoverTriggered: StateFailover,
		EventHeartbeatRestored: StateStable, // could abort
	},
	StateFailover: {
		EventFailoverCompleted: StateRecovered,
	},
	StateRecovered: {
		EventHeartbeatRestored: StateStable,
	},
}

//Personal.AI order the ending
