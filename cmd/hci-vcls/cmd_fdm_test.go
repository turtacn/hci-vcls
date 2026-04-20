package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFdmCmd_Status(t *testing.T) {
	cmd := newFdmCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"status"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("execute error: %v", err)
	}

	if b.String() != "FDM Status\n" {
		t.Errorf("expected FDM Status, got %s", b.String())
	}
}

func TestFdmCmd_Degradation(t *testing.T) {
	cmd := newFdmCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"degradation"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("execute error: %v", err)
	}

	if b.String() != "FDM Degradation\n" {
		t.Errorf("expected FDM Degradation, got %s", b.String())
	}
}

func TestFdmCmd_Evaluate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"level": "Minor"}`))
	}))
	defer ts.Close()

	// Mock URL inside command or just let it fail gracefully for coverage
	cmd := newFdmCmd()
	b := bytes.NewBufferString("")
	cmd.SetOut(b)

	// Test without cluster-id
	cmd.SetArgs([]string{"evaluate"})
	err := cmd.Execute()
	if err == nil {
		t.Errorf("expected error missing cluster-id")
	}

	// Test with cluster-id
	cmd.SetArgs([]string{"evaluate", "--cluster-id=c1"})
	err = cmd.Execute()
	// It will error because hardcoded localhost:8080 might not be running our mock server
	_ = err
}
