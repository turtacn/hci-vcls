package witness

import (
	"context"
	"testing"
)

func TestMemoryClient_Check(t *testing.T) {
	c := NewMemoryClient()
	ctx := context.Background()

	// Default true
	state, err := c.Check(ctx, "vm1")
	if err != nil || !state.Available {
		t.Errorf("Expected available by default")
	}

	// Set explicitly
	c.SetState("vm2", false, "network partition")

	state, err = c.Check(ctx, "vm2")
	if err != nil || state.Available {
		t.Errorf("Expected unavailable, got available")
	}
	if state.Reason != "network partition" {
		t.Errorf("Expected reason 'network partition', got %s", state.Reason)
	}
}

func TestMemoryClient_CheckBatch(t *testing.T) {
	c := NewMemoryClient()
	ctx := context.Background()

	c.SetState("vm1", true, "")
	c.SetState("vm2", false, "down")

	res, err := c.CheckBatch(ctx, []string{"vm1", "vm2", "vm3"})
	if err != nil {
		t.Fatalf("CheckBatch failed: %v", err)
	}

	if len(res) != 3 {
		t.Errorf("Expected 3 results, got %d", len(res))
	}

	if !res["vm1"].Available || res["vm2"].Available || !res["vm3"].Available {
		t.Errorf("Unexpected batch result values: %+v", res)
	}
}

// Personal.AI order the ending
