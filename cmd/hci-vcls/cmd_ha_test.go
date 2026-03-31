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

	// Test tasks
	cmd.SetArgs([]string{"tasks"})
	err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected tasks command to execute, got %v", err)
	}
	if buf.String() == "" {
		t.Errorf("Expected output from tasks command")
	}

	buf.Reset()

	// Test evaluate
	cmd.SetArgs([]string{"evaluate", "100"})
	err = cmd.Execute()
	if err != nil {
		t.Errorf("Expected evaluate command to execute, got %v", err)
	}
	if buf.String() == "" {
		t.Errorf("Expected output from evaluate command")
	}
}

//Personal.AI order the ending