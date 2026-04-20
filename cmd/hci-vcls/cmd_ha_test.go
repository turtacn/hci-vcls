package main

import (
	"bytes"
	"testing"
)

func TestHACmd_Tasks(t *testing.T) {
	cmd := newHACmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"tasks"})

	err := cmd.Execute()
	_ = err // expecting fail on HTTP call
}

func TestHACmd_Evaluate(t *testing.T) {
	cmd := newHACmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"evaluate"})

	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected missing cluster-id error")
	}

	cmd.SetArgs([]string{"evaluate", "--cluster-id=c1"})
	err = cmd.Execute()
	_ = err // expect HTTP fail
}
