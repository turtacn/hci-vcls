package ha

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/qm"
)

type mockQMClient struct {
	qm.Client
	startCalled bool
	startErr    error
}

func (m *mockQMClient) StartVM(ctx context.Context, vmID string, clusterID string, hostID string, bootPath string) (*qm.Task, error) {
	m.startCalled = true
	return &qm.Task{ID: "task-1"}, m.startErr
}

func TestExecutor_Execute(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	mockQM := &mockQMClient{}

	executor := NewExecutor(mockQM, nil, nil, nil, nil, log, 10*time.Millisecond, false)
	ctx := context.Background()

	// Test missing plan
	err := executor.Execute(ctx, nil)
	if err != ErrInvalidPlan {
		t.Errorf("Expected ErrInvalidPlan, got %v", err)
	}

	// Test execution
	plan := &Plan{
		ID:           "plan-1",
		TotalBatches: 1,
		Tasks: []VMTask{
			{ID: "task-1", Status: TaskPending, VMID: "vm-1", TargetHost: "host-2", BatchNo: 1},
			{ID: "task-2", Status: TaskSkipped, BatchNo: 1},
		},
	}
	err = executor.Execute(ctx, plan)
	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}
	if !mockQM.startCalled {
		t.Errorf("Expected QM client StartVM to be called")
	}

	// Test failure
	mockQM.startErr = errors.New("start failed")
	err = executor.Execute(ctx, plan)
	if err != ErrPartialFailure {
		t.Errorf("Expected ErrPartialFailure, got %v", err)
	}

	// Test context cancellation
	ctxCancel, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	err = executor.Execute(ctxCancel, plan)
	if err != ErrLeadershipLost {
		t.Errorf("Expected ErrLeadershipLost, got %v", err)
	}
}

// Personal.AI order the ending
