package ha

import (
	"context"
	"errors"
	"testing"

	"github.com/turtacn/hci-vcls/pkg/cache"
	"github.com/turtacn/hci-vcls/pkg/fdm"
)

type mockCacheManager struct {
	computeMeta *cache.VMComputeMeta
	err         error
}

func (m *mockCacheManager) Start(ctx context.Context) error                         { return nil }
func (m *mockCacheManager) Stop() error                                             { return nil }
func (m *mockCacheManager) GetComputeMeta(ctx context.Context, vmid string) (*cache.VMComputeMeta, error) {
	return m.computeMeta, m.err
}
func (m *mockCacheManager) GetNetworkMeta(ctx context.Context, vmid string) (*cache.VMNetworkMeta, error) {
	return nil, m.err
}
func (m *mockCacheManager) GetStorageMeta(ctx context.Context, vmid string) (*cache.VMStorageMeta, error) {
	return nil, m.err
}
func (m *mockCacheManager) GetHAMeta(ctx context.Context, vmid string) (*cache.VMHAMeta, error) {
	return nil, m.err
}
func (m *mockCacheManager) Sync(ctx context.Context, vmid string) error { return nil }
func (m *mockCacheManager) Stats() cache.CacheStats                     { return cache.CacheStats{} }

type mockFDMAgent struct {
	leaderID string
}

func (m *mockFDMAgent) Start(ctx context.Context) error                         { return nil }
func (m *mockFDMAgent) Stop() error                                             { return nil }
func (m *mockFDMAgent) NodeStates() map[string]fdm.NodeState                    { return nil }
func (m *mockFDMAgent) LocalDegradationLevel() fdm.DegradationLevel             { return fdm.DegradationNone }
func (m *mockFDMAgent) IsLeader() bool                                          { return true }
func (m *mockFDMAgent) LeaderNodeID() string                                    { return m.leaderID }
func (m *mockFDMAgent) OnNodeFailure(callback func(nodeID string))              {}
func (m *mockFDMAgent) OnDegradationChanged(callback func(level fdm.DegradationLevel)) {}
func (m *mockFDMAgent) ClusterView() fdm.ClusterView                            { return fdm.ClusterView{} }

func TestEvaluator_Evaluate(t *testing.T) {
	cacheMgr := &mockCacheManager{
		computeMeta: &cache.VMComputeMeta{VMID: "100", NodeID: "node-1"},
	}
	fdmAgent := &mockFDMAgent{leaderID: "node-1"}

	evaluator := NewEvaluator(cacheMgr, fdmAgent)
	ctx := context.Background()

	cv := ClusterView{
		Nodes: map[string]fdm.NodeState{
			"node-1": fdm.NodeStateAlive,
			"node-2": fdm.NodeStateAlive,
			"node-3": fdm.NodeStateDead,
		},
	}

	decision, err := evaluator.Evaluate(ctx, "100", cv)
	if err != nil {
		t.Fatalf("Expected evaluation to succeed, got %v", err)
	}

	if decision.Action != ActionBoot {
		t.Errorf("Expected ActionBoot, got %v", decision.Action)
	}
	if decision.Path != BootPathMySQL {
		t.Errorf("Expected BootPathMySQL, got %v", decision.Path)
	}
	if decision.TargetNode != "node-2" {
		t.Errorf("Expected target node-2, got %s", decision.TargetNode)
	}
	if decision.Priority != 1 {
		t.Errorf("Expected priority 1, got %d", decision.Priority)
	}

	// Test insufficient resources
	cvOnlySelf := ClusterView{
		Nodes: map[string]fdm.NodeState{
			"node-1": fdm.NodeStateAlive,
			"node-3": fdm.NodeStateDead,
		},
	}

	decision, err = evaluator.Evaluate(ctx, "100", cvOnlySelf)
	if err != ErrInsufficientResources {
		t.Errorf("Expected ErrInsufficientResources, got %v", err)
	}
	if decision.Action != ActionSkip {
		t.Errorf("Expected ActionSkip on error, got %v", decision.Action)
	}
}

func TestBootError(t *testing.T) {
	err := ErrNotLeader
	if err.Error() == "" {
		t.Errorf("Expected error message to not be empty")
	}

	wrappedErr := &BootError{Code: "ERR_CUSTOM", Message: "custom error", Err: errors.New("inner error")}
	if err := wrappedErr.Unwrap(); err == nil {
		t.Errorf("Expected to unwrap an inner error")
	}
}

//Personal.AI order the ending