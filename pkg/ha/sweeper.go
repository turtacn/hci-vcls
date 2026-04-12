package ha

import (
	"context"
	"errors"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/mysql"
)

type SweeperConfig struct {
	StaleThreshold time.Duration
	ScanInterval   time.Duration
}

type Sweeper interface {
	Start(ctx context.Context) error
	Stop() error
	LastRunAt() time.Time
	ReleasedCount() int64
}

type sweeperImpl struct {
	cfg           SweeperConfig
	adapter       mysql.Adapter
	isLeader      func() bool
	metrics       metrics.Metrics
	log           logger.Logger
	lastRunAt     time.Time
	releasedCount int64
	cancel        context.CancelFunc
	done          chan struct{}
}

func NewSweeper(cfg SweeperConfig, adapter mysql.Adapter, isLeader func() bool, m metrics.Metrics, log logger.Logger) Sweeper {
	return &sweeperImpl{
		cfg:      cfg,
		adapter:  adapter,
		isLeader: isLeader,
		metrics:  m,
		log:      log,
		done:     make(chan struct{}),
	}
}

func (s *sweeperImpl) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	go func() {
		defer close(s.done)
		ticker := time.NewTicker(s.cfg.ScanInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.run(ctx)
			}
		}
	}()

	return nil
}

func (s *sweeperImpl) Stop() error {
	if s.cancel != nil {
		s.cancel()
		<-s.done
	}
	return nil
}

func (s *sweeperImpl) LastRunAt() time.Time {
	return s.lastRunAt
}

func (s *sweeperImpl) ReleasedCount() int64 {
	return s.releasedCount
}

func (s *sweeperImpl) run(ctx context.Context) {
	if !s.isLeader() {
		return
	}

	now := time.Now()
	s.lastRunAt = now
	if s.metrics != nil {
		s.metrics.SetSweeperLastRunUnix(float64(now.Unix()))
	}

	thresholdTime := now.Add(-s.cfg.StaleThreshold)
	claims, err := s.adapter.ListStaleBootingClaims(ctx, thresholdTime)
	if err != nil {
		if s.log != nil {
			s.log.Error("Failed to list stale booting claims", "error", err)
		}
		return
	}

	for _, claim := range claims {
		err := s.adapter.ReleaseStaleClaim(ctx, claim.VMID, claim.Token, "stale_sweeper")
		if errors.Is(err, mysql.ErrOptimisticLockFailed) {
			if s.log != nil {
				s.log.Info("Claim already reclaimed, skip", "vmid", claim.VMID, "token", claim.Token)
			}
		} else if err != nil {
			if s.log != nil {
				s.log.Error("Failed to release stale claim", "vmid", claim.VMID, "token", claim.Token, "error", err)
			}
			if s.metrics != nil {
				s.metrics.IncSweeperReleaseFailed()
			}
		} else {
			if s.log != nil {
				s.log.Info("Successfully released stale claim", "vmid", claim.VMID, "token", claim.Token)
			}
			s.releasedCount++
			if s.metrics != nil {
				s.metrics.IncSweeperReleaseOK()
			}
		}
	}
}
