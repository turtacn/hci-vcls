package vcls

import (
	"context"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/cache"
	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/witness"
)

type mockFDMAgent struct {
	cv fdm.ClusterView
}

func (m *mockFDMAgent) Start(ctx context.Context) error                 { return nil }
func (m *mockFDMAgent) Stop() error                                     { return nil }
func (m *mockFDMAgent) LocalDegradationLevel() fdm.DegradationLevel     { return fdm.DegradationNone }
func (m *mockFDMAgent) OnDegradationChanged(func(fdm.DegradationLevel)) {}
func (m *mockFDMAgent) OnNodeFailure(func(string))                      {}
func (m *mockFDMAgent) NodeStates() map[string]fdm.NodeState            { return m.cv.Nodes }
func (m *mockFDMAgent) IsLeader() bool                                  { return true }
func (m *mockFDMAgent) LeaderNodeID() string                            { return "node-1" }
func (m *mockFDMAgent) ClusterView() fdm.ClusterView                    { return m.cv }

func TestService_Refresh(t *testing.T) {
	ctx := context.Background()
	log := logger.NewLogger("debug", "text")
	m := metrics.NewNoopMetrics()

	cfsClient := cfs.NewMemoryClient()
	repo := mysql.NewMemoryVMRepository()
	wit := witness.NewMemoryClient()
	store := NewMemoryStore()
	c := cache.NewMemoryCache[string, bool](1 * time.Minute)

	// Setup data
	cfsClient.AddVM(&cfs.VM{ID: "vm1", ClusterID: "c1", HostID: "h1", PowerState: "running"})
	cfsClient.AddVM(&cfs.VM{ID: "vm2", ClusterID: "c1", HostID: "h2", PowerState: "stopped"})
	cfsClient.AddVM(&cfs.VM{ID: "vm3", ClusterID: "c1", HostID: "h1", PowerState: "running"}) // not protected

	_ = repo.Upsert(ctx, &mysql.VMRecord{VMID: "vm1", ClusterID: "c1", Protected: true})
	_ = repo.Upsert(ctx, &mysql.VMRecord{VMID: "vm2", ClusterID: "c1", Protected: true})

	wit.SetState("vm1", false, "network down")

	fdmAgent := &mockFDMAgent{
		cv: fdm.ClusterView{
			Nodes: map[string]fdm.NodeState{
				"h1": fdm.NodeStateDead,
				"h2": fdm.NodeStateAlive,
			},
		},
	}

	service := NewService(store, cfsClient, repo, wit, fdmAgent, c, m, log)

	// 1. Execute Refresh
	err := service.Refresh(ctx, "c1")
	if err != nil {
		t.Fatalf("Refresh failed: %v", err)
	}

	// List Protected
	prot, _ := service.ListProtected(ctx, "c1")
	if len(prot) != 2 {
		t.Errorf("Expected 2 protected VMs, got %d", len(prot))
	}

	// 2. Protected=false does not appear in EligibleForHA (vm3)
	elig, _ := service.ListEligible(ctx, "c1")
	if len(elig) != 1 {
		t.Errorf("Expected 1 eligible VM, got %d", len(elig))
	}
	if len(elig) == 1 && elig[0].ID != "vm1" {
		t.Errorf("Expected eligible vm to be vm1, got %s", elig[0].ID)
	}

	// 3. HostHealthy=false && Protected=true (vm1) -> Eligible (verified above)

	// 4. Witness mapped
	vm1, _ := service.GetVM(ctx, "vm1")
	if vm1.WitnessAvailable {
		t.Errorf("Expected WitnessAvailable to be false for vm1")
	}

	// 5. Idempotent Refresh caching
	// We call Refresh again, it should return early from cache.
	// If we clear the CFS client, it won't fail because it's cached.
	cfsClient = cfs.NewMemoryClient()                                      // empty client
	service = NewService(store, cfsClient, repo, wit, fdmAgent, c, m, log) // inject empty client but same cache
	err = service.Refresh(ctx, "c1")
	if err != nil {
		t.Errorf("Refresh failed: %v", err)
	}

	// Store should still have the old data because refresh skipped
	elig2, _ := service.ListEligible(ctx, "c1")
	if len(elig2) != 1 {
		t.Errorf("Cache debounce failed, expected 1 eligible VM, got %d", len(elig2))
	}
}

// Personal.AI order the ending
