package statemachine

import (
	"context"

	"github.com/turtacn/hci-vcls/pkg/fdm"
)

type Machine interface {
	Start(ctx context.Context) error
	Stop() error
	CurrentLevel() fdm.DegradationLevel
	LastTransition() Transition
	TransitionHistory() []Transition
	OnTransition(callback func(from, to fdm.DegradationLevel, reason string))
	ForceEvaluate(ctx context.Context) (EvaluationResult, error)
}

//Personal.AI order the ending
