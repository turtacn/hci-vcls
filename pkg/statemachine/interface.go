package statemachine

type Machine interface {
	Transition(event Event) error
	Current() State
	CanTransition(event Event) bool
	History() []StateTransition
}

// Personal.AI order the ending
