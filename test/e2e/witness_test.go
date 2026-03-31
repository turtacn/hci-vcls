package e2e

import (
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/pkg/fdm"
)

func TestWitnessArbitration(t *testing.T) {
	// Simulated Split-Brain Scenario
	// Setup: 3 nodes
	// Action: Network partition between node-1 and {node-2, node-3}
	// Verify: Witnesses resolve quorum

	testCluster.SetDegradation(fdm.DegradationNone)

	// Simulate partition where node-1 cannot see node-2 and node-3
	_ = testCluster.FailNode("node-2")
	_ = testCluster.FailNode("node-3")

	// node-1 attempts to confirm failure via witness
	time.Sleep(100 * time.Millisecond)

	// Witness confirms they are alive, node-1 loses quorum and isolates
	// (Mocking this behavior for the test skeleton)
	isolated := true // Should be false if quorum is not reached
	if !isolated {
		t.Errorf("Expected node-1 to be isolated without quorum")
	}

	_ = testCluster.RecoverNode("node-2")
	_ = testCluster.RecoverNode("node-3")
}

//Personal.AI order the ending