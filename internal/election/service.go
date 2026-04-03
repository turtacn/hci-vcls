package election

import (
	"context"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
)

type ElectionService struct {
	elector Elector
	metrics metrics.Metrics
	log     logger.Logger
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewService(elector Elector, m metrics.Metrics, log logger.Logger) *ElectionService {
	ctx, cancel := context.WithCancel(context.Background())
	return &ElectionService{
		elector: elector,
		metrics: m,
		log:     log,
		ctx:     ctx,
		cancel:  cancel,
	}
}

func (s *ElectionService) Start() error {
	go s.watchLoop()
	return s.elector.Campaign(s.ctx)
}

func (s *ElectionService) Stop() error {
	s.cancel()
	_ = s.elector.Resign(context.Background())
	return s.elector.Close()
}

func (s *ElectionService) watchLoop() {
	watchCh := s.elector.Watch()

	for {
		select {
		case <-s.ctx.Done():
			return
		case status, ok := <-watchCh:
			if !ok {
				return
			}
			result := "lost"
			if status.IsLeader {
				result = "won"
			}
			s.metrics.IncElectionTotal("node", result) // simplifed for mock
			s.log.Info("Election status changed", "isLeader", status.IsLeader, "leaderID", status.LeaderID)
		}
	}
}

// Personal.AI order the ending
