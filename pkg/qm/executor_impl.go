package qm

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type executorImpl struct {
	config QMConfig
}

func NewExecutor(config QMConfig) Executor {
	return &executorImpl{config: config}
}

func (e *executorImpl) Start(ctx context.Context, vmid string, opts BootOptions) BootResult {
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(opts.TimeoutMs)*time.Millisecond)
	defer cancel()

	// Using a mock command for illustration. A real implementation would invoke the `qm` binary.
	cmd := exec.CommandContext(cmdCtx, "echo", "starting", vmid)
	err := cmd.Run()
	if err != nil {
		return BootResult{Success: false, Error: fmt.Errorf("%w: %v", ErrQMTimeout, err)}
	}
	return BootResult{Success: true, Error: nil}
}

func (e *executorImpl) Status(ctx context.Context, vmid string) (VMStatus, error) {
	// A real implementation would parse the output of `qm status <vmid>`
	return VMStatusRunning, nil
}

func (e *executorImpl) Stop(ctx context.Context, vmid string, opts BootOptions) error {
	// ... execution logic
	return nil
}

func (e *executorImpl) Lock(ctx context.Context, vmid string) error {
	return nil
}

func (e *executorImpl) Unlock(ctx context.Context, vmid string) error {
	return nil
}

//Personal.AI order the ending