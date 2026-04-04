package mysql

import (
	"context"
	"testing"
)

func TestVMRepository(t *testing.T) {
	repo := NewMemoryVMRepository()
	ctx := context.Background()

	// 1. Get not found
	_, err := repo.GetByID(ctx, "vm-1")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}

	// 2. Upsert and Get
	record := &VMRecord{
		VMID:        "vm-1",
		ClusterID:   "cluster-1",
		Protected:   true,
		PowerState:  "on",
		CurrentHost: "host-1",
	}

	if err := repo.Upsert(ctx, record); err != nil {
		t.Errorf("Upsert failed: %v", err)
	}

	found, err := repo.GetByID(ctx, "vm-1")
	if err != nil || found.VMID != "vm-1" {
		t.Errorf("GetByID failed: %v", err)
	}

	// 3. List Protected
	record2 := &VMRecord{
		VMID:      "vm-2",
		ClusterID: "cluster-1",
		Protected: false,
	}
	_ = repo.Upsert(ctx, record2)

	protected, err := repo.ListProtected(ctx, "cluster-1")
	if err != nil {
		t.Fatalf("ListProtected failed: %v", err)
	}

	if len(protected) != 1 || protected[0].VMID != "vm-1" {
		t.Errorf("Expected 1 protected VM, got %d", len(protected))
	}
}

func TestHATaskRepository(t *testing.T) {
	repo := NewMemoryHATaskRepository()
	ctx := context.Background()

	task := &HATaskRecord{
		ID:        "task-1",
		VMID:      "vm-1",
		ClusterID: "cluster-1",
		Status:    TaskPending,
	}

	if err := repo.Create(ctx, task); err != nil {
		t.Errorf("Create failed: %v", err)
	}

	if err := repo.UpdateStatus(ctx, "task-1", TaskRunning); err != nil {
		t.Errorf("UpdateStatus failed: %v", err)
	}

	tasks, _ := repo.ListByPlan(ctx, "plan-1")
	if len(tasks) != 1 || tasks[0].Status != TaskRunning {
		t.Errorf("Expected 1 task with RUNNING status, got %d", len(tasks))
	}
}

func TestPlanRepository(t *testing.T) {
	repo := NewMemoryPlanRepository()
	ctx := context.Background()

	plan := &PlanRecord{
		ID:        "plan-1",
		ClusterID: "cluster-1",
		TaskCount: 5,
	}

	if err := repo.Create(ctx, plan); err != nil {
		t.Errorf("Create failed: %v", err)
	}

	found, err := repo.GetByID(ctx, "plan-1")
	if err != nil || found.TaskCount != 5 {
		t.Errorf("GetByID failed: %v", err)
	}

	_, err = repo.GetByID(ctx, "not-exist")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

// Personal.AI order the ending
