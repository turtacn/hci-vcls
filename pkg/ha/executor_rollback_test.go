package ha

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/mysql"
)

func TestExecutor_Finalize_RollbackOnQMStartFailure(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	tx := &mockMySQLTx{}
	mysqlAdp := &mockMySQLAdapter{tx: tx, healthState: mysql.MySQLStateHealthy}
	mockQM := &mockQMClient{startErr: errors.New("qm start failed")}

	executor := NewExecutor(mockQM, nil, mysqlAdp, nil, nil, log, 10*time.Millisecond, false)
	ctx := context.Background()

	plan := &Plan{
		ID:           "plan-1",
		Degradation:  "Minor",
		TotalBatches: 1,
		Tasks:        []VMTask{{ID: "task-1", Status: TaskPending, VMID: "vm-1", BatchNo: 1}},
	}

	err := executor.Execute(ctx, plan)
	if !errors.Is(err, ErrPartialFailure) {
		t.Errorf("expected ErrPartialFailure, got %v", err)
	}

	if !tx.rollbacked {
		t.Errorf("expected tx to be rollbacked on qm start failure")
	}
	if tx.committed {
		t.Errorf("expected tx to NOT be committed on qm start failure")
	}
}

func TestExecutor_Finalize_RollbackFailureDoesNotMaskOriginalError(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	tx := &mockMySQLTx{rollbackErr: errors.New("db down during rollback")}
	mysqlAdp := &mockMySQLAdapter{tx: tx, healthState: mysql.MySQLStateHealthy}
	originalErr := errors.New("qm start failed")
	mockQM := &mockQMClient{startErr: originalErr}

	executor := NewExecutor(mockQM, nil, mysqlAdp, nil, nil, log, 10*time.Millisecond, false)
	ctx := context.Background()

	plan := &Plan{
		ID:           "plan-1",
		Degradation:  "Minor",
		TotalBatches: 1,
		Tasks:        []VMTask{{ID: "task-1", Status: TaskPending, VMID: "vm-1", BatchNo: 1}},
	}

	err := executor.Execute(ctx, plan)
	if !errors.Is(err, ErrPartialFailure) {
		t.Errorf("expected ErrPartialFailure, got %v", err)
	}

	if !tx.rollbacked {
		t.Errorf("expected tx to be rollbacked on qm start failure")
	}

	if plan.Tasks[0].Reason != originalErr.Error() {
		t.Errorf("expected task reason to be %v, got %v", originalErr.Error(), plan.Tasks[0].Reason)
	}
}

func TestExecutor_Finalize_CommitOnSuccess(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	tx := &mockMySQLTx{}
	mysqlAdp := &mockMySQLAdapter{tx: tx, healthState: mysql.MySQLStateHealthy}
	mockQM := &mockQMClient{}

	executor := NewExecutor(mockQM, nil, mysqlAdp, nil, nil, log, 10*time.Millisecond, false)
	ctx := context.Background()

	plan := &Plan{
		ID:           "plan-1",
		Degradation:  "Minor",
		TotalBatches: 1,
		Tasks:        []VMTask{{ID: "task-1", Status: TaskPending, VMID: "vm-1", BatchNo: 1}},
	}

	err := executor.Execute(ctx, plan)
	if err != nil {
		t.Errorf("expected success, got %v", err)
	}

	if !tx.committed {
		t.Errorf("expected tx to be committed on qm start success")
	}
	if tx.rollbacked {
		t.Errorf("expected tx to NOT be rollbacked on qm start success")
	}
}
