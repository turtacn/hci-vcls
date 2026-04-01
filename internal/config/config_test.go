package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Load with an empty config file
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	_ = os.WriteFile(path, []byte(""), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel info, got %s", cfg.LogLevel)
	}
	if cfg.Node.ID != "local-node" {
		t.Errorf("Expected default Node.ID local-node, got %s", cfg.Node.ID)
	}
	if cfg.Server.ListenAddr != ":8080" {
		t.Errorf("Expected default Server.ListenAddr :8080, got %s", cfg.Server.ListenAddr)
	}
}

func TestLoad_Custom(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	yamlContent := `
LogLevel: debug
Node:
  ID: "node-1"
  ClusterID: "prod-cluster"
Server:
  ListenAddr: ":9000"
`
	_ = os.WriteFile(path, []byte(yamlContent), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel debug, got %s", cfg.LogLevel)
	}
	if cfg.Node.ID != "node-1" {
		t.Errorf("Expected Node.ID node-1, got %s", cfg.Node.ID)
	}
	if cfg.Server.ListenAddr != ":9000" {
		t.Errorf("Expected Server.ListenAddr :9000, got %s", cfg.Server.ListenAddr)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	_ = os.WriteFile(path, []byte(""), 0644)

	os.Setenv("VCLS_NODE_ID", "env-node")
	defer os.Unsetenv("VCLS_NODE_ID")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Node.ID != "env-node" {
		t.Errorf("Expected Node.ID env-node, got %s", cfg.Node.ID)
	}
}

func TestLoad_InvalidPath(t *testing.T) {
	// Should fall back to defaults if file doesn't exist
	cfg, err := Load("/tmp/nonexistent_config_vcls_test.yaml")
	if err != nil {
		t.Fatalf("Expected no error when file is missing, got %v", err)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel info, got %s", cfg.LogLevel)
	}
}

//Personal.AI order the ending
