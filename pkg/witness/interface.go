package witness

import "context"

type Client interface {
	Check(ctx context.Context, vmID string) (*WitnessState, error)
	CheckBatch(ctx context.Context, vmIDs []string) (map[string]*WitnessState, error)

	// VoteWeight returns the witness participation weight for quorum.
	// Typically 1 for a single witness node.
	VoteWeight() int

	// ConfirmNodeFailure is called by FDM Leader to get independent
	// confirmation that a peer node is dead (not just network-partitioned).
	ConfirmNodeFailure(ctx context.Context, nodeID string) (bool, error)
}

