package witness

import (
	"context"

	"github.com/turtacn/hci-vcls/internal/logger"
)

type poolImpl struct {
	config WitnessConfig
	log    logger.Logger
}

func NewPool(config WitnessConfig, log logger.Logger) (Pool, error) {
	return &poolImpl{config: config, log: log}, nil
}

func (p *poolImpl) ConfirmFailure(ctx context.Context, req ConfirmationRequest) bool {
	// A minimal confirmation logic simulating reaching out to witnesses to verify.
	p.log.Debug("Confirming failure for node", "nodeID", req.NodeID)
	// Example mock behavior
	return len(p.config.Endpoints) > 0
}

func (p *poolImpl) Quorum(ctx context.Context) bool {
	// Check if a majority of witnesses are accessible
	// Simulated response based on configured endpoints
	return len(p.config.Endpoints) >= 1
}

func (p *poolImpl) Statuses(ctx context.Context) map[string]WitnessStatus {
	statuses := make(map[string]WitnessStatus)
	for _, endpoint := range p.config.Endpoints {
		statuses[endpoint] = StatusHealthy // Mock behavior
	}
	return statuses
}

type adapterImpl struct {
	config WitnessConfig
	log    logger.Logger
}

func NewAdapter(config WitnessConfig, log logger.Logger) (Adapter, error) {
	return &adapterImpl{config: config, log: log}, nil
}

func (a *adapterImpl) Health(ctx context.Context) WitnessStatus {
	// Send ping/health checks
	if len(a.config.Endpoints) > 0 {
		return StatusHealthy
	}
	return StatusUnknown
}

func (a *adapterImpl) ConfirmFailure(ctx context.Context, req ConfirmationRequest) ConfirmationResponse {
	return ConfirmationResponse{Confirmed: true, Error: nil}
}

func (a *adapterImpl) Close() error {
	return nil
}

//Personal.AI order the ending