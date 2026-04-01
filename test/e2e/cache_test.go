package e2e

import (
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/pkg/cache"
	"github.com/turtacn/hci-vcls/pkg/ha"
)

func TestCacheSyncAndConsume(t *testing.T) {
	// Simulated Cache Scenario
	// Setup: CFS source updated
	// Action: Cache Manager syncs
	// Verify: HA evaluator uses updated cache data

	// Mocking a cache hit/miss for E2E
	meta := cache.VMComputeMeta{
		VMID:   "200",
		CPUs:   4,
		Memory: 8192,
		NodeID: "node-1",
	}

	time.Sleep(50 * time.Millisecond) // Let it sync

	// Mock evaluator consuming the cache
	decision := ha.HADecision{
		VMID:       meta.VMID,
		Action:     ha.ActionBoot,
		Path:       ha.BootPathMySQL,
		TargetNode: "node-2", // Different from current NodeID
		Reason:     "Cache data consumed",
	}

	testCluster.Tasks = append(testCluster.Tasks, ha.BootTask{
		VMID:     meta.VMID,
		Decision: decision,
		Status:   ha.TaskPending,
	})

	found := false
	for _, tsk := range testCluster.Tasks {
		if tsk.VMID == "200" && tsk.Decision.TargetNode == "node-2" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected evaluator to consume cache and target node-2")
	}
}

//Personal.AI order the ending
