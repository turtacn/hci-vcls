package ha

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/mysql"
)

type mockAdapter struct {
	mysql.Adapter
	listClaimsFunc func() ([]mysql.BootClaim, error)
	releaseFunc    func() error
}

func (m *mockAdapter) ListStaleBootingClaims(ctx context.Context, threshold time.Time) ([]mysql.BootClaim, error) {
	if m.listClaimsFunc != nil {
		return m.listClaimsFunc()
	}
	return nil, nil
}

func (m *mockAdapter) ReleaseStaleClaim(ctx context.Context, vmid, token string, reason string) error {
	if m.releaseFunc != nil {
		return m.releaseFunc()
	}
	return nil
}

func TestSweeper_ReleasesStaleBootingClaim(t *testing.T) {
	released := 0
	adapter := &mockAdapter{
		listClaimsFunc: func() ([]mysql.BootClaim, error) {
			return []mysql.BootClaim{
				{VMID: "vm1", Token: "token1"},
			}, nil
		},
		releaseFunc: func() error {
			released++
			return nil
		},
	}
	s := NewSweeper(SweeperConfig{}, adapter, func() bool { return true }, metrics.NewNoopMetrics(), logger.NewLogger("debug", "console"))

	s.(*sweeperImpl).run(context.Background())

	if released != 1 {
		t.Errorf("expected 1 release, got %d", released)
	}
}

func TestSweeper_SkipsWhenNotLeader(t *testing.T) {
	calls := 0
	adapter := &mockAdapter{
		listClaimsFunc: func() ([]mysql.BootClaim, error) {
			calls++
			return nil, nil
		},
	}
	s := NewSweeper(SweeperConfig{}, adapter, func() bool { return false }, metrics.NewNoopMetrics(), logger.NewLogger("debug", "console"))

	s.(*sweeperImpl).run(context.Background())

	if calls != 0 {
		t.Errorf("expected 0 calls, got %d", calls)
	}
}

func TestSweeper_OptimisticLockFailure_NotAnError(t *testing.T) {
	adapter := &mockAdapter{
		listClaimsFunc: func() ([]mysql.BootClaim, error) {
			return []mysql.BootClaim{
				{VMID: "vm1", Token: "token1"},
			}, nil
		},
		releaseFunc: func() error {
			return mysql.ErrOptimisticLockFailed
		},
	}
	s := NewSweeper(SweeperConfig{}, adapter, func() bool { return true }, metrics.NewNoopMetrics(), logger.NewLogger("debug", "console"))

	s.(*sweeperImpl).run(context.Background())

	// Should not panic or crash, logging handles the "error" as info
	if s.ReleasedCount() != 0 {
		t.Errorf("Expected 0 released, got %d", s.ReleasedCount())
	}
}

func TestSweeper_DBDown_DoesNotPanic(t *testing.T) {
	adapter := &mockAdapter{
		listClaimsFunc: func() ([]mysql.BootClaim, error) {
			return nil, errors.New("db connection failed")
		},
	}
	s := NewSweeper(SweeperConfig{}, adapter, func() bool { return true }, metrics.NewNoopMetrics(), logger.NewLogger("debug", "console"))

	s.(*sweeperImpl).run(context.Background())

	// Should not panic
}

func TestSweeper_StopIsIdempotent(t *testing.T) {
	s := NewSweeper(SweeperConfig{ScanInterval: 1 * time.Second}, &mockAdapter{}, func() bool { return true }, metrics.NewNoopMetrics(), logger.NewLogger("debug", "console"))
	_ = s.Start(context.Background())

	err1 := s.Stop()
	err2 := s.Stop()

	if err1 != nil || err2 != nil {
		t.Errorf("Expected Stop to be idempotent, got %v, %v", err1, err2)
	}
}
