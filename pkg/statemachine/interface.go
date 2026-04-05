package statemachine

type Machine interface {
	Transition(event Event) error
	Current() State
	CanTransition(event Event) bool
	History() []StateTransition

	// Adapters for agent interface to avoid cyclic imports
	TransitionString(event string) error
	EvaluateWithInput(input interface{}) (string, string)
	CurrentLevel() string
}

// Personal.AI order the ending
