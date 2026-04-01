package qm

import "context"

type MockExecutor struct {
	StartFunc  func(ctx context.Context, vmid string, opts BootOptions) BootResult
	StatusFunc func(ctx context.Context, vmid string) (VMStatus, error)
	StopFunc   func(ctx context.Context, vmid string, opts BootOptions) error
	LockFunc   func(ctx context.Context, vmid string) error
	UnlockFunc func(ctx context.Context, vmid string) error
}

func (m *MockExecutor) Start(ctx context.Context, vmid string, opts BootOptions) BootResult {
	if m.StartFunc != nil {
		return m.StartFunc(ctx, vmid, opts)
	}
	return BootResult{Success: true}
}

func (m *MockExecutor) Status(ctx context.Context, vmid string) (VMStatus, error) {
	if m.StatusFunc != nil {
		return m.StatusFunc(ctx, vmid)
	}
	return VMStatusRunning, nil
}

func (m *MockExecutor) Stop(ctx context.Context, vmid string, opts BootOptions) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx, vmid, opts)
	}
	return nil
}

func (m *MockExecutor) Lock(ctx context.Context, vmid string) error {
	if m.LockFunc != nil {
		return m.LockFunc(ctx, vmid)
	}
	return nil
}

func (m *MockExecutor) Unlock(ctx context.Context, vmid string) error {
	if m.UnlockFunc != nil {
		return m.UnlockFunc(ctx, vmid)
	}
	return nil
}

//Personal.AI order the ending
