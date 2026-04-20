package main

import (
	"bytes"
	"testing"
)

import (
	"os"
)

func TestExecuteCommand(t *testing.T) {
	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	rootCmd.SetArgs([]string{})

	err := Execute()
	if err != nil {
		t.Errorf("execute error: %v", err)
	}
}

func TestInitConfig(t *testing.T) {
	cfgFile = "does-not-exist.yaml"
	initConfig()
	// it will print an error and continue
}

func TestMain_Exit(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"hci-vcls", "help"}
	main()
}
