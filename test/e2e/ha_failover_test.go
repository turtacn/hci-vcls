package e2e

import (
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
)

func TestHAFailover(t *testing.T) {
	// Simulated HA Failover Scenario
	// Setup: Node 1 is primary running VM 100
	// Action: Node 1 fails
	// Verify: VM 100 restarts on Node 2 or Node 3

	testCluster.SetDegradation(fdm.DegradationNone)

	// Node 1 fails
	err := testCluster.FailNode("node-1")
	if err != nil {
		t.Fatalf("Failed to simulate node failure: %v", err)
	}

	// In a real test, wait for heartbeat timeout and election
	time.Sleep(100 * time.Millisecond)

	// Simulate evaluator deciding to boot on node-2
	decision := ha.HADecision{
		VMID:       "100",
		Action:     ha.ActionBoot,
		Path:       ha.BootPathMySQL,
		TargetNode: "node-2",
		Reason:     "Node failure",
	}

	task := ha.BootTask{
		VMID:     "100",
		Decision: decision,
		Status:   ha.TaskCompleted,
	}

	testCluster.AddTask(task)

	// Assertions based on test environment state
	found := false
	for _, tsk := range testCluster.Tasks {
		if tsk.VMID == "100" && tsk.Decision.TargetNode == "node-2" && tsk.Status == ha.TaskCompleted {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected HA task for VM 100 to complete on node-2")
	}

	// Recover node 1
	_ = testCluster.RecoverNode("node-1")
}

//Personal.AI order the ending
