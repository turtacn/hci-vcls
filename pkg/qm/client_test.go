package qm

import (
	"context"
	"testing"
	"time"
)

func TestMemoryClient_StartAndWait_Success(t *testing.T) {
	c := NewMemoryClient(0.0, 10*time.Millisecond) // 0% failure rate
	ctx := context.Background()

	task, err := c.StartVM(ctx, "vm1", "c1", "host1", "mysql")
	if err != nil {
		t.Fatalf("StartVM failed: %v", err)
	}

	if task.Status != TaskPending {
		t.Errorf("Expected TaskPending, got %v", task.Status)
	}

	finalTask, err := c.WaitTask(ctx, task.ID, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitTask failed: %v", err)
	}

	if finalTask.Status != TaskDone {
		t.Errorf("Expected TaskDone, got %v", finalTask.Status)
	}
}

func TestMemoryClient_StartAndWait_Failure(t *testing.T) {
	c := NewMemoryClient(1.0, 10*time.Millisecond) // 100% failure rate
	ctx := context.Background()

	task, _ := c.StartVM(ctx, "vm1", "c1", "host1", "mysql")

	finalTask, err := c.WaitTask(ctx, task.ID, 200*time.Millisecond)
	if err != nil {
		t.Fatalf("WaitTask failed: %v", err)
	}

	if finalTask.Status != TaskFailed {
		t.Errorf("Expected TaskFailed, got %v", finalTask.Status)
	}
}

func TestMemoryClient_StopVM(t *testing.T) {
	c := NewMemoryClient(0.0, 10*time.Millisecond)
	ctx := context.Background()

	task, _ := c.StopVM(ctx, "vm1", "c1")
	finalTask, _ := c.WaitTask(ctx, task.ID, 200*time.Millisecond)

	if finalTask.Status != TaskDone {
		t.Errorf("Expected TaskDone, got %v", finalTask.Status)
	}
}

// Personal.AI order the ending
