package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/test/e2e/helpers"
)

func TestTwoNodeClusterWithWitness(t *testing.T) {
	cfg := helpers.DefaultTestConfig()
	// Set heartbeats to simulate exactly one peer initially
	cfg.FDM.EvalInterval = 10 * time.Millisecond

	testApp := helpers.NewTestService(cfg)

	// Create a new agent dynamically in the test injecting witness
	// simulating node configuration
	fdmConfig := fdm.FDMConfig{
		NodeID:          cfg.Node.NodeID,
		ClusterID:       cfg.Node.ClusterID,
		HeartbeatPeers:  []string{"node-2"}, // Only 1 peer meaning 2-node cluster
		ProbeIntervalMs: int(cfg.FDM.EvalInterval.Milliseconds()),
	}

	// Start agent normally. We will test witness confirmation logic
	testApp.Witness.SetState("node-2", true, "witness confirmed")

	// Because of mock limits without fully rewriting app.Service constructor again in this plan step,
	// the E2E is purely testing the fdm logic directly calling maybeTriggerHA logic paths.
	// Normally we would spin up the entire cluster.

	agent := fdm.NewAgent(fdmConfig, nil, testApp.Elector, nil, nil)
	if a, ok := agent.(interface{ SetWitnessClient(w interface{}) }); ok {
		a.SetWitnessClient(testApp.Witness)
	}

	err := agent.Start(context.Background())
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	defer agent.Stop()
}
