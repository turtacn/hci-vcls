package witness

import (
	"context"
	"sync"
	"time"
)

type MemoryClient struct {
	mu     sync.RWMutex
	states map[string]*WitnessState
}

var _ Client = &MemoryClient{}

func NewMemoryClient() *MemoryClient {
	return &MemoryClient{
		states: make(map[string]*WitnessState),
	}
}

// For test injection
func (c *MemoryClient) SetState(vmID string, available bool, reason string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.states[vmID] = &WitnessState{
		VMID:      vmID,
		Available: available,
		Reason:    reason,
		CheckedAt: time.Now(),
	}
}

func (c *MemoryClient) Check(ctx context.Context, vmID string) (*WitnessState, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	state, ok := c.states[vmID]
	if !ok {
		// Default to available if not configured for tests
		return &WitnessState{
			VMID:      vmID,
			Available: true,
			CheckedAt: time.Now(),
		}, nil
	}

	return state, nil
}

func (c *MemoryClient) CheckBatch(ctx context.Context, vmIDs []string) (map[string]*WitnessState, error) {
	result := make(map[string]*WitnessState)
	for _, vmID := range vmIDs {
		state, err := c.Check(ctx, vmID)
		if err != nil {
			return nil, err
		}
		result[vmID] = state
	}
	return result, nil
}

func (c *MemoryClient) VoteWeight() int {
	return 1
}

func (c *MemoryClient) ConfirmNodeFailure(ctx context.Context, nodeID string) (bool, error) {
	// Simple mock: assume node is always failed if confirmed through witness in tests
	return true, nil
}
