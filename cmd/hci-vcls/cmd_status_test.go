package main

import (
	"bytes"
	"testing"
)

func TestStatusCmd(t *testing.T) {
	cmd := newStatusCmd()
	if cmd.Use != "status" {
		t.Errorf("Expected use 'status', got %s", cmd.Use)
	}

	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Expected status command to execute, got %v", err)
	}
	if buf.String() == "" {
		t.Errorf("Expected output from status command")
	}
}

//Personal.AI order the ending
