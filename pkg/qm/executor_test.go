package qm

import (
	"context"
	"errors"
	"testing"
)

func TestMockExecutor_Success(t *testing.T) {
	mock := &MockExecutor{}

	ctx := context.Background()
	opts := BootOptions{TimeoutMs: 100}

	result := mock.Start(ctx, "100", opts)
	if !result.Success {
		t.Errorf("Expected Start to succeed by default")
	}

	status, err := mock.Status(ctx, "100")
	if err != nil {
		t.Errorf("Expected no error from Status, got %v", err)
	}
	if status != VMStatusRunning {
		t.Errorf("Expected VMStatusRunning, got %v", status)
	}

	err = mock.Stop(ctx, "100", opts)
	if err != nil {
		t.Errorf("Expected no error from Stop, got %v", err)
	}

	err = mock.Lock(ctx, "100")
	if err != nil {
		t.Errorf("Expected no error from Lock, got %v", err)
	}

	err = mock.Unlock(ctx, "100")
	if err != nil {
		t.Errorf("Expected no error from Unlock, got %v", err)
	}
}

func TestMockExecutor_CustomBehavior(t *testing.T) {
	expectedErr := errors.New("custom start error")

	mock := &MockExecutor{
		StartFunc: func(ctx context.Context, vmid string, opts BootOptions) BootResult {
			return BootResult{Success: false, Error: expectedErr}
		},
		StatusFunc: func(ctx context.Context, vmid string) (VMStatus, error) {
			return VMStatusUnknown, errors.New("vm unknown")
		},
		StopFunc: func(ctx context.Context, vmid string, opts BootOptions) error {
			return errors.New("stop failed")
		},
		LockFunc: func(ctx context.Context, vmid string) error {
			return errors.New("lock failed")
		},
		UnlockFunc: func(ctx context.Context, vmid string) error {
			return errors.New("unlock failed")
		},
	}

	ctx := context.Background()
	opts := BootOptions{TimeoutMs: 100}

	result := mock.Start(ctx, "100", opts)
	if result.Success {
		t.Errorf("Expected Start to fail")
	}
	if result.Error != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, result.Error)
	}

	status, err := mock.Status(ctx, "100")
	if err == nil {
		t.Errorf("Expected error from Status")
	}
	if status != VMStatusUnknown {
		t.Errorf("Expected VMStatusUnknown, got %v", status)
	}

	err = mock.Stop(ctx, "100", opts)
	if err == nil {
		t.Errorf("Expected error from Stop")
	}

	err = mock.Lock(ctx, "100")
	if err == nil {
		t.Errorf("Expected error from Lock")
	}

	err = mock.Unlock(ctx, "100")
	if err == nil {
		t.Errorf("Expected error from Unlock")
	}
}

func TestQMError(t *testing.T) {
	err := ErrVMNotFound
	if err.Error() == "" {
		t.Errorf("Expected error message to not be empty")
	}

	wrappedErr := &QMError{Code: "ERR_CUSTOM", Message: "custom error", Err: errors.New("inner error")}
	if err := wrappedErr.Unwrap(); err == nil {
		t.Errorf("Expected to unwrap an inner error")
	}
}

//Personal.AI order the ending