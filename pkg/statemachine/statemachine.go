package statemachine

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrIllegalTransition = errors.New("statemachine: illegal transition")
)

type machineImpl struct {
	mu      sync.RWMutex
	current State
	history []StateTransition
}

var _ Machine = &machineImpl{}

func NewMachine() Machine {
	return &machineImpl{
		current: StateInit,
		history: make([]StateTransition, 0),
	}
}

// 迁移矩阵（完整定义，非法事件返回 error，不 panic，不改变状态）：
func (m *machineImpl) getNextState(current State, event Event) (State, bool) {
	switch current {
	case StateInit:
		if event == EventHeartbeatRestored {
			return StateStable, true
		}
	case StateStable:
		if event == EventDegradationDetected || event == EventHeartbeatLost {
			return StateDegraded, true
		}
	case StateDegraded:
		if event == EventEvaluationStarted {
			return StateEvaluating, true
		}
		if event == EventHeartbeatRestored {
			return StateStable, true
		}
	case StateEvaluating:
		if event == EventFailoverTriggered {
			return StateFailover, true
		}
		if event == EventHeartbeatRestored {
			return StateStable, true
		}
	case StateFailover:
		if event == EventFailoverCompleted {
			return StateRecovered, true
		}
	case StateRecovered:
		if event == EventHeartbeatRestored {
			return StateStable, true
		}
	}
	return current, false
}

func (m *machineImpl) Transition(event Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	nextState, valid := m.getNextState(m.current, event)
	if !valid {
		return ErrIllegalTransition
	}

	m.history = append(m.history, StateTransition{
		From:      m.current,
		To:        nextState,
		Event:     event,
		Timestamp: time.Now(),
	})
	m.current = nextState

	return nil
}

func (m *machineImpl) Current() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

func (m *machineImpl) CanTransition(event Event) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, valid := m.getNextState(m.current, event)
	return valid
}

func (m *machineImpl) History() []StateTransition {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h := make([]StateTransition, len(m.history))
	copy(h, m.history)
	return h
}

// Personal.AI order the ending
