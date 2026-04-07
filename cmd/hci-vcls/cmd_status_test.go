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
}

