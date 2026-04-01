package mysql

import (
	"errors"
	"testing"
)

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

//Personal.AI order the ending
