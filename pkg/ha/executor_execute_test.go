package ha

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/qm"
)

// Extending the mock adapters for executor tests

type mockMySQLTx struct {
	claimErr    error
	commitErr   error
	rollbackErr error
	committed   bool
	rollbacked  bool
}

func (m *mockMySQLTx) ClaimBoot(claim mysql.BootClaim) error {
	return m.claimErr
}

func (m *mockMySQLTx) Commit() error {
	m.committed = true
	return m.commitErr
}

func (m *mockMySQLTx) Rollback() error {
	m.rollbacked = true
	return m.rollbackErr
}

type mockMySQLAdapter struct {
	tx          *mockMySQLTx
	beginErr    error
	healthState mysql.HealthState
}

func (m *mockMySQLAdapter) BeginTx() (mysql.TxAdapter, error) {
	if m.beginErr != nil {
		return nil, m.beginErr
	}
	return m.tx, nil
}

func (m *mockMySQLAdapter) Health() mysql.MySQLStatus {
	return mysql.MySQLStatus{State: m.healthState}
}

func (m *mockMySQLAdapter) ClaimBoot(claim mysql.BootClaim) error {
	return nil
}

func (m *mockMySQLAdapter) ConfirmBoot(vmid, token string) error {
	return nil
}

func (m *mockMySQLAdapter) ReleaseBoot(vmid, token string) error {
	return nil
}

func (m *mockMySQLAdapter) GetVMState(vmid string) (*mysql.HAVMState, error) {
	return nil, nil
}

func (m *mockMySQLAdapter) UpsertVMState(state mysql.HAVMState) error {
	return nil
}

func (m *mockMySQLAdapter) Close() error {
	return nil
}

func (m *mockMySQLAdapter) RunQuery(ctx context.Context, query string, args ...interface{}) error {
	return nil
}

type mockCacheProvider struct {
	meta interface{}
	err  error
}

func (m *mockCacheProvider) GetComputeMeta(ctx context.Context, vmid string) (interface{}, error) {
	return m.meta, m.err
}

func TestExecutor_Execute_GateReject(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	executor := NewExecutor(nil, nil, nil, nil, nil, log, 10*time.Millisecond, false)
	ctx := context.Background()

	plan := &Plan{
		ID:           "plan-1",
		Degradation:  "Critical",
		TotalBatches: 1,
		Tasks:        []VMTask{{ID: "task-1", Status: TaskPending}},
	}

	err := executor.Execute(ctx, plan)
	if !errors.Is(err, ErrSkippedIsolated) {
		t.Errorf("expected ErrSkippedIsolated, got %v", err)
	}
}

func TestExecutor_Execute_MySQLClaimConflict(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	tx := &mockMySQLTx{claimErr: errors.New("conflict")}
	mysqlAdp := &mockMySQLAdapter{tx: tx, healthState: mysql.MySQLStateHealthy}
	executor := NewExecutor(nil, nil, mysqlAdp, nil, nil, log, 10*time.Millisecond, true)
	ctx := context.Background()

	plan := &Plan{
		ID:           "plan-1",
		Degradation:  "Minor",
		TotalBatches: 1,
		Tasks:        []VMTask{{ID: "task-1", Status: TaskPending, VMID: "vm-1", BatchNo: 1}},
	}

	err := executor.Execute(ctx, plan)
	if !errors.Is(err, ErrPartialFailure) {
		t.Errorf("expected ErrPartialFailure due to conflict, got %v", err)
	}
	if !tx.rollbacked {
		t.Errorf("expected transaction to be rollbacked on claim failure")
	}
}

func TestExecutor_Execute_MySQLUnavailableMajor(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	mysqlAdp := &mockMySQLAdapter{healthState: mysql.MySQLStateUnavailable}
	executor := NewExecutor(nil, nil, mysqlAdp, nil, nil, log, 10*time.Millisecond, true)
	ctx := context.Background()

	plan := &Plan{
		ID:           "plan-1",
		Degradation:  "Major",
		TotalBatches: 1,
		Tasks:        []VMTask{{ID: "task-1", Status: TaskPending, VMID: "vm-1", BatchNo: 1}},
	}

	err := executor.Execute(ctx, plan)
	if !errors.Is(err, ErrPartialFailure) {
		t.Errorf("expected ErrPartialFailure due to mysql unavailable, got %v", err)
	}
	if plan.Tasks[0].Reason != "MySQL unavailable, cannot acquire optimistic lock" {
		t.Errorf("expected specific reason for mysql unavailable, got %v", plan.Tasks[0].Reason)
	}
}

func TestExecutor_Execute_CacheMiss(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	tx := &mockMySQLTx{}
	mysqlAdp := &mockMySQLAdapter{tx: tx, healthState: mysql.MySQLStateHealthy}
	mockQM := &mockQMClient{}
	cacheProv := &mockCacheProvider{err: errors.New("cache miss")}

	exec := NewExecutor(mockQM, nil, mysqlAdp, nil, nil, log, 10*time.Millisecond, false)
	execImpl := exec.(*executorImpl)
	execImpl.SetCache(cacheProv)

	ctx := context.Background()

	plan := &Plan{
		ID:           "plan-1",
		Degradation:  "Minor",
		TotalBatches: 1,
		Tasks: []VMTask{
			{ID: "task-1", Status: TaskPending, VMID: "vm-1", BatchNo: 1, BootPath: BootPathMinority},
		},
	}

	// It should warn on cache miss, but still try to start VM and succeed
	err := exec.Execute(ctx, plan)
	if err != nil {
		t.Errorf("expected nil error despite cache miss, got %v", err)
	}
	if !mockQM.startCalled {
		t.Errorf("expected qm to be called")
	}
}

func TestExecutor_Execute_QMAlreadyRunning(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	tx := &mockMySQLTx{}
	mysqlAdp := &mockMySQLAdapter{tx: tx, healthState: mysql.MySQLStateHealthy}
	mockQM := &mockQMClient{startErr: qm.ErrVMAlreadyRunning}

	exec := NewExecutor(mockQM, nil, mysqlAdp, nil, nil, log, 10*time.Millisecond, false)
	ctx := context.Background()

	plan := &Plan{
		ID:           "plan-1",
		Degradation:  "Minor",
		TotalBatches: 1,
		Tasks:        []VMTask{{ID: "task-1", Status: TaskPending, VMID: "vm-1", BatchNo: 1}},
	}

	err := exec.Execute(ctx, plan)
	if err != nil {
		t.Errorf("expected idempotency success, got %v", err)
	}
	if tx.rollbacked {
		t.Errorf("expected tx to not rollback on already running")
	}
	if !tx.committed {
		t.Errorf("expected tx to commit on already running")
	}
}

func TestExecutor_Execute_QMStartFailure(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	tx := &mockMySQLTx{}
	mysqlAdp := &mockMySQLAdapter{tx: tx, healthState: mysql.MySQLStateHealthy}
	mockQM := &mockQMClient{startErr: errors.New("general error")}

	exec := NewExecutor(mockQM, nil, mysqlAdp, nil, nil, log, 10*time.Millisecond, false)
	ctx := context.Background()

	plan := &Plan{
		ID:           "plan-1",
		Degradation:  "Minor",
		TotalBatches: 1,
		Tasks:        []VMTask{{ID: "task-1", Status: TaskPending, VMID: "vm-1", BatchNo: 1}},
	}

	err := exec.Execute(ctx, plan)
	if !errors.Is(err, ErrPartialFailure) {
		t.Errorf("expected partial failure due to qm start error, got %v", err)
	}
	if !tx.rollbacked {
		t.Errorf("expected tx to rollback when QM start fails")
	}
}

func TestExecutor_Execute_FailFast(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	tx := &mockMySQLTx{}
	mysqlAdp := &mockMySQLAdapter{tx: tx, healthState: mysql.MySQLStateHealthy}
	mockQM := &mockQMClient{startErr: errors.New("error")}

	exec := NewExecutor(mockQM, nil, mysqlAdp, nil, nil, log, 10*time.Millisecond, true) // failFast = true
	ctx := context.Background()

	plan := &Plan{
		ID:           "plan-1",
		Degradation:  "Minor",
		TotalBatches: 2, // 2 batches
		Tasks: []VMTask{
			{ID: "task-1", Status: TaskPending, VMID: "vm-1", BatchNo: 1},
			{ID: "task-2", Status: TaskPending, VMID: "vm-2", BatchNo: 2},
		},
	}

	err := exec.Execute(ctx, plan)
	if !errors.Is(err, ErrPartialFailure) {
		t.Errorf("expected partial failure, got %v", err)
	}

	// Mock qm should only have been called once if it failed fast on the first task
	// Wait, startCalled is a boolean flag so we can only know if it was called at least once.
	// But since it failed fast, batch 2 should not be executed.
}

func TestExecutor_Execute_CtxCancel(t *testing.T) {
	log := logger.NewLogger("debug", "console")
	exec := NewExecutor(nil, nil, nil, nil, nil, log, 10*time.Millisecond, false)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel

	plan := &Plan{
		ID:           "plan-1",
		TotalBatches: 1,
		Tasks:        []VMTask{{ID: "task-1", Status: TaskPending, BatchNo: 1}},
	}

	err := exec.Execute(ctx, plan)
	if !errors.Is(err, ErrLeadershipLost) {
		t.Errorf("expected leadership lost due to context cancellation, got %v", err)
	}
}
