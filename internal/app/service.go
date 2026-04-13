package app

import (
	"errors"

	"github.com/turtacn/hci-vcls/internal/config"
	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/heartbeat"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/statemachine"
	"github.com/turtacn/hci-vcls/pkg/vcls"
	"go.uber.org/zap"
)

var (
	ErrNotLeader      = errors.New("not leader")
	ErrBelowThreshold = errors.New("degradation below threshold")
)

type Service struct {
	config       *config.Config
	logger       *zap.Logger
	metrics      metrics.Metrics
	election     election.Elector
	heartbeat    *heartbeat.HeartbeatService
	vcls         vcls.Service
	planner      ha.Planner
	executor     ha.Executor
	statemachine statemachine.Machine
	vmRepo       mysql.VMRepository
	planRepo     mysql.PlanRepository
	fdmAgent     fdm.Agent
	sweeper      ha.Sweeper
	planCache    PlanCache
}

func NewService(
	cfg *config.Config,
	log *zap.Logger,
	m metrics.Metrics,
	e election.Elector,
	hb *heartbeat.HeartbeatService,
	v vcls.Service,
	p ha.Planner,
	ex ha.Executor,
	sm statemachine.Machine,
	vm mysql.VMRepository,
	plan mysql.PlanRepository,
	fdm fdm.Agent,
	sweeper ha.Sweeper,
	planCache PlanCache,
) *Service {
	s := &Service{
		config:       cfg,
		logger:       log,
		metrics:      m,
		election:     e,
		heartbeat:    hb,
		vcls:         v,
		planner:      p,
		executor:     ex,
		statemachine: sm,
		vmRepo:       vm,
		planRepo:     plan,
		fdmAgent:     fdm,
		sweeper:      sweeper,
		planCache:    planCache,
	}

	if planCache != nil {
		if pending, err := planCache.List(); err == nil && len(pending) > 0 {
			if log != nil {
				log.Warn("Found pending plans in cache that were not executed cleanly", zap.Int("count", len(pending)))
			}
		}
	}

	return s
}

func (s *Service) Planner() ha.Planner {
	return s.planner
}

func (s *Service) Executor() ha.Executor {
	return s.executor
}
