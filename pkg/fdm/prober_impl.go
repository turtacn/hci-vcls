package fdm

import (
	"context"

	"github.com/turtacn/hci-vcls/internal/heartbeat"
	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/witness"
)

type proberImpl struct {
	hb      heartbeat.Heartbeater
	pool    witness.Pool
	log     logger.Logger
	metrics metrics.Metrics
}

func NewProber(hb heartbeat.Heartbeater, pool witness.Pool, log logger.Logger, m metrics.Metrics) Prober {
	return &proberImpl{hb: hb, pool: pool, log: log, metrics: m}
}

func (p *proberImpl) ProbeL0(ctx context.Context) ProbeResult {
	// e.g., ping hardware
	return ProbeResult{Level: HeartbeatL0, Success: true}
}

func (p *proberImpl) ProbeL1(ctx context.Context) ProbeResult {
	// e.g., host OS level checks
	return ProbeResult{Level: HeartbeatL1, Success: true}
}

func (p *proberImpl) ProbeL2(ctx context.Context) ProbeResult {
	// e.g., checking cluster interconnects
	return ProbeResult{Level: HeartbeatL2, Success: true}
}

func (p *proberImpl) ProbeAll(ctx context.Context) map[HeartbeatLevel]ProbeResult {
	return map[HeartbeatLevel]ProbeResult{
		HeartbeatL0: p.ProbeL0(ctx),
		HeartbeatL1: p.ProbeL1(ctx),
		HeartbeatL2: p.ProbeL2(ctx),
	}
}

//Personal.AI order the ending
