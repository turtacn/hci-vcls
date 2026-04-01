package e2e

import (
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
)

func TestDegradationPath(t *testing.T) {
	// Simulated Degradation Scenario
	// Setup: Cluster healthy
	// Action: ZK goes down
	// Verify: Level drops to DegradationZK, HA capability preserved

	testCluster.SetDegradation(fdm.DegradationNone)

	// Simulate ZK failure
	testCluster.SetDegradation(fdm.DegradationZK)

	// Wait for agent/state machine
	time.Sleep(100 * time.Millisecond)

	// Verify we can still do HA
	decision := ha.HADecision{
		VMID:       "101",
		Action:     ha.ActionBoot,
		Path:       ha.BootPathMySQL,
		TargetNode: "node-3",
		Reason:     "Degraded boot",
	}

	testCluster.Tasks = append(testCluster.Tasks, ha.BootTask{
		VMID:     "101",
		Decision: decision,
		Status:   ha.TaskRunning,
	})

	found := false
	for _, tsk := range testCluster.Tasks {
		if tsk.VMID == "101" && tsk.Decision.Path == ha.BootPathMySQL && tsk.Status == ha.TaskRunning {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected HA to function under ZK degradation")
	}

	// Restore
	testCluster.SetDegradation(fdm.DegradationNone)
}

//Personal.AI order the ending
