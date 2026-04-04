package statemachine

import "fmt"

type TransitionError struct {
	Code    string
	Message string
	Err     error
}

func (e *TransitionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("statemachine error %s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("statemachine error %s: %s", e.Code, e.Message)
}

func (e *TransitionError) Unwrap() error {
	return e.Err
}

var (
	ErrMachineNotStarted = &TransitionError{Code: "ERR_MACHINE_NOT_STARTED", Message: "state machine not started"}
	ErrInvalidTransition = &TransitionError{Code: "ERR_INVALID_TRANSITION", Message: "invalid state transition"}
	ErrEvaluationTimeout = &TransitionError{Code: "ERR_EVALUATION_TIMEOUT", Message: "evaluation timeout"}
	ErrCooldownActive    = &TransitionError{Code: "ERR_COOLDOWN_ACTIVE", Message: "transition cooldown active"}
)

// Personal.AI order the ending
