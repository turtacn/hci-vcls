package main

import (
	"context"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/test/e2e/helpers"
)


func TestBuildServer(t *testing.T) {
	cfg := helpers.DefaultTestConfig()

	cfg.DataDir = t.TempDir()
	svc, srv, hb, err := buildServer(cfg)
	if err != nil {
		t.Fatalf("buildServer failed: %v", err)
	}
	if svc == nil || srv == nil || hb == nil {
		t.Fatalf("buildServer returned nil components")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_ = ctx
}

func TestServeCmd(t *testing.T) {
	// Not fully running runServe to avoid blocking on signal.Notify, just verify the command is mapped.
	if serveCmd.Use != "serve" {
		t.Errorf("expected command to be serve")
	}
}

func TestDummyProber(t *testing.T) {
	prober := &dummyProber{}
	ctx := context.Background()
	if !prober.ProbeL0(ctx).Success {
		t.Errorf("expected true")
	}
	if !prober.ProbeL1(ctx).Success {
		t.Errorf("expected true")
	}
	if !prober.ProbeL2(ctx).Success {
		t.Errorf("expected true")
	}
	if !prober.ProbeAll(ctx)[fdm.HeartbeatL0].Success {
		t.Errorf("expected true")
	}
}
