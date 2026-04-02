package app

import (
	"errors"

	"go.uber.org/zap"
	"github.com/turtacn/hci-vcls/internal/config"
	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/heartbeat"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/statemachine"
	"github.com/turtacn/hci-vcls/pkg/vcls"
)

var (
	ErrNotLeader       = errors.New("not leader")
	ErrBelowThreshold  = errors.New("degradation below threshold")
)

type Service struct {
	config       *config.Config
	logger       *zap.Logger
	metrics      metrics.Metrics
	election     election.Elector
	heartbeat    heartbeat.Heartbeater
	vcls         vcls.Service
	planner      ha.Planner
	executor     ha.Executor
	statemachine statemachine.Machine
	vmRepo       mysql.VMRepository
	planRepo     mysql.PlanRepository
	fdmAgent     fdm.Agent
}

func NewService(
	cfg *config.Config,
	log *zap.Logger,
	m metrics.Metrics,
	e election.Elector,
	hb heartbeat.Heartbeater,
	v vcls.Service,
	p ha.Planner,
	ex ha.Executor,
	sm statemachine.Machine,
	vm mysql.VMRepository,
	plan mysql.PlanRepository,
	fdm fdm.Agent,
) *Service {
	return &Service{
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
	}
}

// Personal.AI order the ending
