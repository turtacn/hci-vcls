package witness

import "context"

type Adapter interface {
	Health(ctx context.Context) WitnessStatus
	ConfirmFailure(ctx context.Context, req ConfirmationRequest) ConfirmationResponse
	Close() error
}

type Pool interface {
	ConfirmFailure(ctx context.Context, req ConfirmationRequest) bool
	Quorum(ctx context.Context) bool
	Statuses(ctx context.Context) map[string]WitnessStatus
}

//Personal.AI order the ending
