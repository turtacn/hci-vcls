package main

import (
	"bytes"
	"testing"
)

func TestVersionCmd(t *testing.T) {
	b := bytes.NewBufferString("")
	versionCmd.SetOut(b)
	versionCmd.SetArgs([]string{})
	err := versionCmd.Execute()

	if err != nil {
		t.Errorf("version execution error: %v", err)
	}

}
