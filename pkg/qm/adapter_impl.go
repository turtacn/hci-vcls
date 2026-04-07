package qm

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type adapterImpl struct {
	qmPath string
}

func NewQMAdapter(qmPath string) *adapterImpl {
	if qmPath == "" {
		qmPath = "/usr/sbin/qm"
	}
	return &adapterImpl{qmPath: qmPath}
}

func (a *adapterImpl) Start(ctx context.Context, vmid string, opts BootOptions) BootResult {
	// First check status for idempotency
	status, _ := a.Status(ctx, vmid)
	if status == VMStatusRunning {
		return BootResult{Success: true, Error: nil}
	}

	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(opts.TimeoutMs)*time.Millisecond)
	defer cancel()

	args := []string{"start", vmid}

	if opts.TimeoutMs > 0 {
		// Mock config inject
		args = append(args, "--timeout", fmt.Sprintf("%d", opts.TimeoutMs))
	}

	cmd := exec.CommandContext(cmdCtx, a.qmPath, args...)
	out, err := cmd.CombinedOutput()

	if err != nil {
		outStr := string(out)
		// Check if it's already running according to the output message
		if strings.Contains(outStr, "is already running") {
			return BootResult{Success: true, Error: nil}
		}
		return BootResult{Success: false, Error: fmt.Errorf("qm start failed: %w output: %s", err, outStr)}
	}

	return BootResult{Success: true, Error: nil}
}

func (a *adapterImpl) Status(ctx context.Context, vmid string) (VMStatus, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, a.qmPath, "status", vmid)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return VMStatusUnknown, err
	}

	outStr := string(out)
	if strings.Contains(outStr, "status: running") {
		return VMStatusRunning, nil
	}
	if strings.Contains(outStr, "status: stopped") {
		return VMStatusStopped, nil
	}

	return VMStatusUnknown, nil
}

func (a *adapterImpl) Stop(ctx context.Context, vmid string, opts BootOptions) error {
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(opts.TimeoutMs)*time.Millisecond)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, a.qmPath, "stop", vmid)
	return cmd.Run()
}

func (a *adapterImpl) Lock(ctx context.Context, vmid string) error {
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// qm set vmid -lock ha
	cmd := exec.CommandContext(cmdCtx, a.qmPath, "set", vmid, "-lock", "ha")
	return cmd.Run()
}

func (a *adapterImpl) Unlock(ctx context.Context, vmid string) error {
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(cmdCtx, a.qmPath, "unlock", vmid)
	return cmd.Run()
}

