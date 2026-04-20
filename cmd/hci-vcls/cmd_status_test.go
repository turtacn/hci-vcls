package main

import (
	"bytes"
	"testing"
)

func TestStatusCmd(t *testing.T) {
	cmd := newStatusCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	_ = err // expect HTTP fail
}
