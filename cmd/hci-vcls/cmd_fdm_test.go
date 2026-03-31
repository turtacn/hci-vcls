package main

import (
	"bytes"
	"testing"
)

func TestFDMCmd(t *testing.T) {
	cmd := newFdmCmd()
	if cmd.Use != "fdm" {
		t.Errorf("Expected use 'fdm', got %s", cmd.Use)
	}

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// Test status
	cmd.SetArgs([]string{"status"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected status command to execute, got %v", err)
	}
	if buf.String() == "" {
		t.Errorf("Expected output from status command")
	}

	buf.Reset()

	// Test degradation
	cmd.SetArgs([]string{"degradation"})
	err = cmd.Execute()
	if err != nil {
		t.Errorf("Expected degradation command to execute, got %v", err)
	}
	if buf.String() == "" {
		t.Errorf("Expected output from degradation command")
	}
}

//Personal.AI order the ending