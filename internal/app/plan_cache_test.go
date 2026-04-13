package app

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/turtacn/hci-vcls/pkg/ha"
)

func TestFSPlanCache_PutGetDeleteList(t *testing.T) {
	dir := t.TempDir()

	cache, err := NewFSPlanCache(dir)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %v", err)
	}

	plan := &ha.Plan{
		ID:        "plan-1",
		ClusterID: "c1",
		Tasks: []ha.VMTask{
			{ID: "task-1", VMID: "vm1", BootPath: "minority"},
		},
	}

	err = cache.Put(plan)
	if err != nil {
		t.Fatalf("unexpected error on put: %v", err)
	}

	path := filepath.Join(dir, "plan-1.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("expected plan file to exist: %s", path)
	}

	p, err := cache.Get("plan-1")
	if err != nil {
		t.Fatalf("unexpected error on get: %v", err)
	}
	if p.ID != plan.ID || p.ClusterID != plan.ClusterID || len(p.Tasks) != 1 {
		t.Errorf("plan mismatch: got %+v", p)
	}

	plans, err := cache.List()
	if err != nil {
		t.Fatalf("unexpected error on list: %v", err)
	}
	if len(plans) != 1 {
		t.Errorf("expected 1 plan, got %d", len(plans))
	}

	err = cache.Delete("plan-1")
	if err != nil {
		t.Fatalf("unexpected error on delete: %v", err)
	}

	_, err = cache.Get("plan-1")
	if err == nil {
		t.Fatalf("expected error on get after delete")
	}

	plans, err = cache.List()
	if err != nil {
		t.Fatalf("unexpected error on list after delete: %v", err)
	}
	if len(plans) != 0 {
		t.Errorf("expected 0 plans, got %d", len(plans))
	}
}
