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
	if cfg.LogFormat != "text" {
		t.Errorf("Expected default LogFormat text, got %s", cfg.LogFormat)
	}
	if cfg.ZK.SessionTimeoutMs != 5000 {
		t.Errorf("Expected default ZK timeout 5000, got %d", cfg.ZK.SessionTimeoutMs)
	}
	if cfg.HA.MaxConcurrentBoots != 5 {
		t.Errorf("Expected default MaxConcurrentBoots 5, got %d", cfg.HA.MaxConcurrentBoots)
	}
}

func TestLoad_Custom(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	yamlContent := `
LogLevel: debug
LogFormat: json
ZK:
  Endpoints:
    - 192.168.1.1:2181
  SessionTimeoutMs: 10000
HA:
  MaxConcurrentBoots: 10
`
	_ = os.WriteFile(path, []byte(yamlContent), 0644)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.LogLevel != "debug" {
		t.Errorf("Expected LogLevel debug, got %s", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("Expected LogFormat json, got %s", cfg.LogFormat)
	}
	if cfg.ZK.SessionTimeoutMs != 10000 {
		t.Errorf("Expected ZK timeout 10000, got %d", cfg.ZK.SessionTimeoutMs)
	}
	if len(cfg.ZK.Endpoints) != 1 || cfg.ZK.Endpoints[0] != "192.168.1.1:2181" {
		t.Errorf("Expected ZK endpoints ['192.168.1.1:2181'], got %v", cfg.ZK.Endpoints)
	}
	if cfg.HA.MaxConcurrentBoots != 10 {
		t.Errorf("Expected MaxConcurrentBoots 10, got %d", cfg.HA.MaxConcurrentBoots)
	}
}

func TestLoad_InvalidPath(t *testing.T) {
	// Should fall back to defaults if file doesn't exist
	cfg, err := Load("/tmp/nonexistent_config.yaml")
	if err != nil {
		t.Fatalf("Expected no error when file is missing, got %v", err)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("Expected default LogLevel info, got %s", cfg.LogLevel)
	}
}

//Personal.AI order the ending