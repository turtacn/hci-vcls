package zk

import (
	"errors"
	"testing"

	"github.com/turtacn/hci-vcls/internal/logger"
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

func TestAdapterImpl_FailurePaths(t *testing.T) {
	// Use 127.0.0.1 with dummy port to avoid dns lookup failure, trigger timeout
	cfg := ZKConfig{Endpoints: []string{"127.0.0.1:12345"}, SessionTimeoutMs: 10}

	// Since go-zookeeper connects async, it should return an adapter structure
	adapter, err := NewAdapter(cfg, logger.Default())
	if err != nil {
		t.Fatalf("expected NewAdapter to return object but err %v", err)
	}

	h := adapter.Health()
	if h.State != ZKStateUnavailable {
		t.Errorf("Expected unavailable health, got %v", h.State)
	}

	h = adapter.IsReadOnly()
	if h.State != ZKStateHealthy {
		t.Errorf("Expected healthy for read only stub, got %v", h.State)
	}

	h = adapter.Ping()
	if h.State != ZKStateUnavailable {
		t.Errorf("Expected unavailable for ping, got %v", h.State)
	}

	_ = adapter.Close()
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

