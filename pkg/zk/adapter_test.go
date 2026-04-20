package zk

import (
	"errors"
	"testing"
)

type mockAdapter struct {
	healthState HealthState
	isReadOnly  HealthState
	pingState   HealthState
	err         error
	closed      bool
}

func (m *mockAdapter) Health() ZKStatus {
	return ZKStatus{State: m.healthState, Error: m.err}
}

func (m *mockAdapter) IsReadOnly() ZKStatus {
	return ZKStatus{State: m.isReadOnly, Error: m.err}
}

func (m *mockAdapter) Ping() ZKStatus {
	return ZKStatus{State: m.pingState, Error: m.err}
}

func (m *mockAdapter) Close() error {
	m.closed = true
	return nil
}

func TestZKAdapter(t *testing.T) {
	mock := &mockAdapter{
		healthState: ZKStateHealthy,
		isReadOnly:  ZKStateHealthy,
		pingState:   ZKStateHealthy,
	}

	status := mock.Health()
	if status.State != ZKStateHealthy {
		t.Errorf("Expected healthy, got %v", status.State)
	}

	mock.healthState = ZKStateUnavailable
	mock.err = errors.New("connection failed")
	status = mock.Health()
	if status.State != ZKStateUnavailable {
		t.Errorf("Expected unavailable, got %v", status.State)
	}

	status = mock.Ping()
	if status.State != ZKStateHealthy {
		t.Errorf("Expected healthy, got %v", status.State)
	}

	status = mock.IsReadOnly()
	if status.State != ZKStateHealthy {
		t.Errorf("Expected healthy for read only, got %v", status.State)
	}

	err := mock.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
	if !mock.closed {
		t.Errorf("Expected closed flag to be set")
	}
}

func TestHealthStateString(t *testing.T) {
	tests := []struct {
		state    HealthState
		expected string
	}{
		{ZKStateHealthy, "Healthy"},
		{ZKStateReadOnly, "ReadOnly"},
		{ZKStateUnavailable, "Unavailable"},
		{HealthState(99), "Unknown"},
	}

	for _, tt := range tests {
		if tt.state.String() != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, tt.state.String())
		}
	}
}

