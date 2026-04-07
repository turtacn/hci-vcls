package witness

import "context"

type Client interface {
	Check(ctx context.Context, vmID string) (*WitnessState, error)
	CheckBatch(ctx context.Context, vmIDs []string) (map[string]*WitnessState, error)
}

