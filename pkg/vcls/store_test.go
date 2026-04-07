package vcls

import (
	"testing"
	"time"
)

func TestMemoryStore(t *testing.T) {
	store := NewMemoryStore()

	vm1 := &VM{ID: "101", ClusterID: "cluster1", Protected: true, EligibleForHA: true, LastSyncAt: time.Now()}
	vm2 := &VM{ID: "102", ClusterID: "cluster1", Protected: true, EligibleForHA: false, LastSyncAt: time.Now()}
	vm3 := &VM{ID: "103", ClusterID: "cluster2", Protected: true, EligibleForHA: true, LastSyncAt: time.Now()}

	store.Put(vm1)
	store.Put(vm2)
	store.Put(vm3)

	vm, ok := store.Get("101")
	if !ok || vm.ID != "101" {
		t.Errorf("expected to find vm 101, got %v", vm)
	}

	_, ok = store.Get("nonexistent")
	if ok {
		t.Error("expected not to find vm nonexistent")
	}

	list := store.List("cluster1")
	if len(list) != 2 {
		t.Errorf("expected 2 vms in cluster1, got %d", len(list))
	}

	eligible := store.ListEligible("cluster1")
	if len(eligible) != 1 {
		t.Errorf("expected 1 eligible vm in cluster1, got %d", len(eligible))
	}

	status := store.Status("cluster1")
	if status.VMCount != 2 {
		t.Errorf("expected 2 VMCount, got %d", status.VMCount)
	}
	if status.ProtectedCount != 2 {
		t.Errorf("expected 2 ProtectedCount, got %d", status.ProtectedCount)
	}
	if status.EligibleCount != 1 {
		t.Errorf("expected 1 EligibleCount, got %d", status.EligibleCount)
	}
	if status.LastRefreshAt.IsZero() {
		t.Error("expected LastRefreshAt to be set")
	}
}
