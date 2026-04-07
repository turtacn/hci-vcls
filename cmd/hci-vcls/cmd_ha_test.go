package main

import (
	"bytes"
	"testing"
)

func TestHACmd(t *testing.T) {
	cmd := newHACmd()
	if cmd.Use != "ha" {
		t.Errorf("Expected use 'ha', got %s", cmd.Use)
	}

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Since we mock API, just verify structure
	cmd.SetArgs([]string{"evaluate"})
	err := cmd.Execute()
	if err == nil {
		t.Errorf("Expected evaluate command to fail without cluster-id")
	}
}

