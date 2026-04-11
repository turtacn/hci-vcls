package app

import (
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/statemachine"
)

// stateMachineAdapter wraps statemachine.Machine to implement ha.StateProvider.
type stateMachineAdapter struct {
	sm statemachine.Machine
}

// NewStateMachineAdapter creates a new StateProvider adapter around statemachine.Machine.
func NewStateMachineAdapter(sm statemachine.Machine) ha.StateProvider {
	return &stateMachineAdapter{sm: sm}
}

// CurrentLevel maps the statemachine's current level to a string format expected by ha.Executor.
func (a *stateMachineAdapter) CurrentLevel() string {
	return a.sm.CurrentLevel()
}
