package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoad_DefaultsPass(t *testing.T) {
	// Need minimum required fields to pass validation
	os.Setenv("HCI_NODE_NODE_ID", "test-node")
	os.Setenv("HCI_NODE_CLUSTER_ID", "test-cluster")
	defer os.Unsetenv("HCI_NODE_NODE_ID")
	defer os.Unsetenv("HCI_NODE_CLUSTER_ID")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Expected defaults to pass validation, got %v", err)
	}

	if cfg.Heartbeat.Interval != 5*time.Second {
		t.Errorf("Expected default heartbeat interval 5s, got %v", cfg.Heartbeat.Interval)
	}
	if cfg.HA.BatchSize != 5 {
		t.Errorf("Expected default batch size 5, got %d", cfg.HA.BatchSize)
	}
}

func TestLoad_HeartbeatIntervalInvalid(t *testing.T) {
	os.Setenv("HCI_NODE_NODE_ID", "test-node")
	os.Setenv("HCI_NODE_CLUSTER_ID", "test-cluster")
	os.Setenv("HCI_HEARTBEAT_INTERVAL", "-1s")
	defer os.Clearenv()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	_, err := Load(path)
	if err == nil {
		t.Fatalf("Expected validation to fail for invalid heartbeat interval")
	}

	if !strings.Contains(err.Error(), "Interval") || !strings.Contains(err.Error(), "gt") {
		t.Errorf("Expected explicit error about interval, got %v", err)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	os.Setenv("HCI_NODE_NODE_ID", "custom-node")
	os.Setenv("HCI_NODE_CLUSTER_ID", "test-cluster")
	defer os.Clearenv()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Node.NodeID != "custom-node" {
		t.Errorf("Expected NodeID to be 'custom-node', got '%s'", cfg.Node.NodeID)
	}
}

func TestLoad_MissingClusterID(t *testing.T) {
	os.Setenv("HCI_NODE_NODE_ID", "test-node")
	// Intentionally omitting HCI_NODE_CLUSTER_ID
	defer os.Clearenv()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	_, err := Load(path)
	if err == nil {
		t.Fatalf("Expected validation to fail due to missing cluster_id")
	}

	if !strings.Contains(err.Error(), "ClusterID") || !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected explicit error about missing ClusterID, got %v", err)
	}
}

// Personal.AI order the ending
