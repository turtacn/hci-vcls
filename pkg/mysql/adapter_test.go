package mysql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/turtacn/hci-vcls/internal/logger"
)

func TestAdapterImpl_SQLMock(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	adapter := &adapterImpl{db: db, log: logger.Default()}
	ctx := context.Background()

	// Health
	mock.ExpectPing().WillReturnError(nil)
	status := adapter.Health()
	if status.State != MySQLStateHealthy {
		t.Errorf("expected healthy, got %v", status.State)
	}

	// BeginTx
	mock.ExpectBegin()
	txAdapter, err := adapter.BeginTx()
	if err != nil {
		t.Errorf("expected no error on BeginTx, got %v", err)
	}

	// Tx ClaimBoot
	mock.ExpectExec("UPDATE ha_vm_state").
		WithArgs("node1", "token1", "100").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = txAdapter.ClaimBoot(BootClaim{TargetNode: "node1", Token: "token1", VMID: "100"})
	if err != nil {
		t.Errorf("expected no error on tx ClaimBoot, got %v", err)
	}

	mock.ExpectCommit()
	_ = txAdapter.Commit()

	// Adapter ClaimBoot
	mock.ExpectExec("UPDATE ha_vm_state").
		WithArgs("node2", "token2", "200").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = adapter.ClaimBoot(BootClaim{TargetNode: "node2", Token: "token2", VMID: "200"})
	if err != nil {
		t.Errorf("expected no error on ClaimBoot, got %v", err)
	}

	// ConfirmBoot
	_ = adapter.ConfirmBoot("100", "token1")

	// ReleaseBoot
	_ = adapter.ReleaseBoot("100", "token1")

	// GetVMState
	state, err := adapter.GetVMState("100")
	if err != nil || state.Status != "running" {
		t.Errorf("expected no error and state running, got %v", state)
	}

	// UpsertVMState
	_ = adapter.UpsertVMState(HAVMState{VMID: "100", Token: "token1", TargetNode: "node1", Status: "running"})

	// ListStaleBootingClaims
	staleRows := sqlmock.NewRows([]string{"vmid", "token", "target_node", "updated_at"}).
		AddRow("100", "token1", "node1", time.Now())
	mock.ExpectQuery("SELECT vmid, token, target_node, updated_at FROM ha_vm_state").
		WillReturnRows(staleRows)

	claims, err := adapter.ListStaleBootingClaims(ctx, time.Now())
	if err != nil || len(claims) != 1 {
		t.Errorf("expected 1 claim, got %v", err)
	}

	// ReleaseStaleClaim
	mock.ExpectExec("UPDATE ha_vm_state").
		WithArgs("reason", "100", "token1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	_ = adapter.ReleaseStaleClaim(ctx, "100", "token1", "reason")

	// Close
	mock.ExpectClose()
	_ = adapter.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

type mockAdapter struct {
	healthState HealthState
	err         error
	closed      bool
	vmState     *HAVMState
}

func (m *mockAdapter) Health() MySQLStatus {
	return MySQLStatus{State: m.healthState, Error: m.err}
}

func (m *mockAdapter) ClaimBoot(claim BootClaim) error {
	return m.err
}

func (m *mockAdapter) ConfirmBoot(vmid, token string) error {
	return m.err
}

func (m *mockAdapter) ReleaseBoot(vmid, token string) error {
	return m.err
}

func (m *mockAdapter) GetVMState(vmid string) (*HAVMState, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.vmState, nil
}

func (m *mockAdapter) UpsertVMState(state HAVMState) error {
	m.vmState = &state
	return m.err
}

func (m *mockAdapter) Close() error {
	m.closed = true
	return m.err
}

func TestAdapterImpl_FailurePaths(t *testing.T) {
	// Attempt connecting to invalid DSN to trigger failures natively
	cfg := MySQLConfig{DSN: "invalid-user:pass@tcp(127.0.0.1:0)/fake"}
	adapter, err := NewAdapter(cfg, nil)
	if err != nil {
		t.Fatalf("expected NewAdapter to return object but err %v", err)
	}

	h := adapter.Health()
	if h.State != MySQLStateUnavailable {
		t.Errorf("Expected unavailable health, got %v", h.State)
	}

	_, err = adapter.BeginTx()
	if err == nil {
		t.Errorf("Expected BeginTx to fail")
	}

	err = adapter.ClaimBoot(BootClaim{})
	if err == nil {
		t.Errorf("Expected ClaimBoot to fail")
	}

	_ = adapter.ConfirmBoot("100", "token")
	_ = adapter.ReleaseBoot("100", "token")
	_, _ = adapter.GetVMState("100")
	_ = adapter.UpsertVMState(HAVMState{})

	_, err = adapter.ListStaleBootingClaims(context.Background(), time.Now())
	if err == nil {
		t.Errorf("Expected ListStaleBootingClaims to fail")
	}

	err = adapter.ReleaseStaleClaim(context.Background(), "100", "token", "reason")
	if err == nil {
		t.Errorf("Expected ReleaseStaleClaim to fail")
	}

	_ = adapter.Close()
}

func TestMockAdapter_ClaimBoot(t *testing.T) {
	mock := &mockAdapter{err: nil}

	err := mock.ClaimBoot(BootClaim{VMID: "100"})
	if err != nil {
		t.Errorf("Expected nil error on ClaimBoot, got %v", err)
	}

	err = mock.ConfirmBoot("100", "t1")
	if err != nil {
		t.Errorf("Expected nil error on ConfirmBoot, got %v", err)
	}

	err = mock.ReleaseBoot("100", "t1")
	if err != nil {
		t.Errorf("Expected nil error on ReleaseBoot, got %v", err)
	}
}

func TestMySQLAdapterMock(t *testing.T) {
	mock := &mockAdapter{
		healthState: MySQLStateHealthy,
		vmState:     &HAVMState{VMID: "100", Token: "t1", TargetNode: "node1", Status: "running"},
	}

	status := mock.Health()
	if status.State != MySQLStateHealthy {
		t.Errorf("Expected healthy, got %v", status.State)
	}

	state, err := mock.GetVMState("100")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if state.TargetNode != "node1" {
		t.Errorf("Expected node1, got %s", state.TargetNode)
	}

	err = mock.UpsertVMState(HAVMState{VMID: "101", Token: "t2", TargetNode: "node2", Status: "stopped"})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if mock.vmState.TargetNode != "node2" {
		t.Errorf("Expected node2, got %s", mock.vmState.TargetNode)
	}

	err = mock.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
	if !mock.closed {
		t.Errorf("Expected closed flag to be true")
	}

	mock.healthState = MySQLStateUnavailable
	mock.err = errors.New("db error")
	status = mock.Health()
	if status.State != MySQLStateUnavailable {
		t.Errorf("Expected unavailable, got %v", status.State)
	}
}

func TestHealthStateString(t *testing.T) {
	tests := []struct {
		state    HealthState
		expected string
	}{
		{MySQLStateHealthy, "Healthy"},
		{MySQLStateReadOnly, "ReadOnly"},
		{MySQLStateUnavailable, "Unavailable"},
		{HealthState(99), "Unknown"},
	}

	for _, tt := range tests {
		if tt.state.String() != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, tt.state.String())
		}
	}
}

func TestMySQLError(t *testing.T) {
	err := ErrBootTokenConflict
	if err.Error() == "" {
		t.Errorf("Expected error message to not be empty")
	}

	wrappedErr := &MySQLError{Code: "ERR_CUSTOM", Message: "custom error", Err: errors.New("inner error")}
	if err := wrappedErr.Unwrap(); err == nil {
		t.Errorf("Expected to unwrap an inner error")
	}
}

