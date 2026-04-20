package qm

import (
	"context"
	"testing"
)

func TestQMAdapter_Status_Unknown(t *testing.T) {
	// A basic test that hits the code path where 'qm' doesn't exist
	adapter := NewQMAdapter("/does/not/exist/qm_fake_path")
	ctx := context.Background()

	status, err := adapter.Status(ctx, "100")
	if err == nil {
		t.Errorf("Expected error from fake qm path")
	}
	if status != VMStatusUnknown {
		t.Errorf("Expected status unknown")
	}

	// Test start failure fallback
	res := adapter.Start(ctx, "100", BootOptions{TimeoutMs: 10})
	if res.Success {
		t.Errorf("Expected failure on Start")
	}

	// Test stop failure fallback
	err = adapter.Stop(ctx, "100", BootOptions{TimeoutMs: 10})
	if err == nil {
		t.Errorf("Expected error on Stop")
	}

	// Test lock/unlock fallback
	err = adapter.Lock(ctx, "100")
	if err == nil {
		t.Errorf("Expected error on Lock")
	}
	err = adapter.Unlock(ctx, "100")
	if err == nil {
		t.Errorf("Expected error on Unlock")
	}
}
