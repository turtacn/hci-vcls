package statemachine

type StateTransitionMachine interface {
	Current() State
	Event(e Event) error
}

//Personal.AI order the ending
