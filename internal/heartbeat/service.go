package heartbeat

import (
	"context"
	"time"

	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/statemachine"
)

type HeartbeatService struct {
	config    HeartbeatConfig
	heartbeater Heartbeater
	monitor   Monitor
	elector   election.Elector
	evaluator fdm.Evaluator
	sm        statemachine.Machine
	metrics   metrics.Metrics
	log       logger.Logger
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewService(config HeartbeatConfig, hb Heartbeater, monitor Monitor, elector election.Elector, evaluator fdm.Evaluator, sm statemachine.Machine, m metrics.Metrics, log logger.Logger) *HeartbeatService {
	ctx, cancel := context.WithCancel(context.Background())
	s := &HeartbeatService{
		config:      config,
		heartbeater: hb,
		monitor:     monitor,
		elector:     elector,
		evaluator:   evaluator,
		sm:          sm,
		metrics:     m,
		log:         log,
		ctx:         ctx,
		cancel:      cancel,
	}

	if hb != nil {
		hb.OnDigestReceived(s.OnDigestReceived)
	}

	return s
}

func (s *HeartbeatService) Start() error {
	go s.loop()
	return nil
}

func (s *HeartbeatService) Stop() error {
	s.cancel()
	return nil
}

func (s *HeartbeatService) OnDigestReceived(digest StateDigest) {
	if s.elector != nil {
		s.elector.ReceivePeerState(digest.NodeID, digest.Term, digest.CandidateID, digest.IsLeader)
	}
}

func (s *HeartbeatService) loop() {
	ticker := time.NewTicker(time.Duration(s.config.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case t := <-ticker.C:
			s.monitor.CheckTimeouts(t, time.Duration(s.config.TimeoutMs)*time.Millisecond)

			// Simple mock evaluation loop
			summaries := s.monitor.ListSummaries("cluster-1") // mocked cluster id for now

			var hosts []fdm.HostState
			var allHealthy = true
			for _, sum := range summaries {
				if !sum.Healthy {
					allHealthy = false
					if s.metrics != nil {
						s.metrics.IncHeartbeatLost(sum.NodeID, sum.ClusterID)
					}
				}
				hosts = append(hosts, fdm.HostState{
					NodeID:        sum.NodeID,
					ClusterID:     sum.ClusterID,
					Healthy:       sum.Healthy,
					LostCount:     sum.LostCount,
					LastHeartbeat: sum.LastSeenAt,
				})
			}

			if s.sm != nil {
				if allHealthy {
					_ = s.sm.Transition(statemachine.EventHeartbeatRestored)
				} else {
					_ = s.sm.Transition(statemachine.EventHeartbeatLost)
				}
			}

			// Update elector digest with our state before it broadcasts via heartbeater
			if s.elector != nil && s.heartbeater != nil {
				term, voteFor, isLeader := s.elector.CurrentTermAndVote()
				s.heartbeater.UpdateDigest(term, voteFor, isLeader)
			}

			// Only evaluate if leader
			if s.elector != nil && s.elector.IsLeader() && s.evaluator != nil {
				state, err := s.evaluator.Evaluate(s.ctx, "cluster-1", s.elector.Status().LeaderID, hosts, true)
				if err != nil {
					if s.log != nil {
						s.log.Error("FDM Evaluation failed", "error", err)
					}
				} else {
					if s.log != nil {
						s.log.Debug("FDM Evaluated", "degradation", state.Degradation, "reason", state.Reason)
					}
					if state.Degradation != fdm.DegradationNone {
						if s.sm != nil {
							_ = s.sm.Transition(statemachine.EventDegradationDetected)
						}
						if s.metrics != nil {
							s.metrics.SetDegradationLevel(state.ClusterID, 1.0) // simplified mock
						}
					} else {
						if s.metrics != nil {
							s.metrics.SetDegradationLevel(state.ClusterID, 0)
						}
					}
				}
			}
		}
	}
}

