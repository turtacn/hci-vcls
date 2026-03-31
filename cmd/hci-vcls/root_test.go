package main

import (
	"bytes"
	"testing"
)

func TestRootCmd(t *testing.T) {
	// A simple test to ensure root command can be executed without panicking
	// Normally we would capture output, but executing without args just prints help
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Expected root command to execute successfully, got %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Errorf("Expected help output, got empty string")
	}
}

func TestInitConfig_MissingFile(t *testing.T) {
	// Temporarily modify the package variable
	originalCfgFile := cfgFile
	defer func() { cfgFile = originalCfgFile }()

	cfgFile = "/tmp/nonexistent_config.yaml"
	initConfig()

	if appConfig == nil {
		t.Errorf("Expected default config to be loaded even if file is missing")
	}
}

//Personal.AI order the ending