package cfs

import (
	"path/filepath"
	"testing"
)

func TestCFSAdapter(t *testing.T) {
	dir := t.TempDir()
	config := CFSConfig{MountPath: dir, TimeoutMs: 100}

	// This relies on the real implementation
	adapter, err := NewAdapter(config, nil)
	if err != nil {
		t.Fatalf("Failed to create adapter: %v", err)
	}

	status := adapter.Health()
	if status.State != CFSStateHealthy {
		t.Errorf("Expected healthy, got %v", status.State)
	}

	status = adapter.IsWritable()
	if status.State != CFSStateHealthy {
		t.Errorf("Expected writable, got %v", status.State)
	}

	err = adapter.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}

	// Test unmounted directory
	invalidDir := filepath.Join(dir, "nonexistent")
	invalidConfig := CFSConfig{MountPath: invalidDir, TimeoutMs: 100}
	invalidAdapter, _ := NewAdapter(invalidConfig, nil)

	status = invalidAdapter.Health()
	if status.State != CFSStateUnmounted {
		t.Errorf("Expected unmounted state, got %v", status.State)
	}
}

func TestHealthStateString(t *testing.T) {
	tests := []struct {
		state    HealthState
		expected string
	}{
		{CFSStateHealthy, "Healthy"},
		{CFSStateReadOnly, "ReadOnly"},
		{CFSStateUnmounted, "Unmounted"},
		{CFSStateUnavailable, "Unavailable"},
		{HealthState(99), "Unknown"},
	}

	for _, tt := range tests {
		if tt.state.String() != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, tt.state.String())
		}
	}
}

//Personal.AI order the ending