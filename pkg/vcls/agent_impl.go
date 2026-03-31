package vcls

import (
	"context"

	"github.com/turtacn/hci-vcls/pkg/fdm"
)

type agentImpl struct {
	config       VCLSConfig
	agent        fdm.Agent
	currentState ClusterServiceState
	activeCaps   []Capability
	cbs          []func(fdm.DegradationLevel)
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewAgent(config VCLSConfig, agent fdm.Agent) Agent {
	ctx, cancel := context.WithCancel(context.Background())
	return &agentImpl{
		config:       config,
		agent:        agent,
		currentState: ServiceStateHealthy,
		activeCaps:   make([]Capability, 0),
		cbs:          make([]func(fdm.DegradationLevel), 0),
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (a *agentImpl) Start(ctx context.Context) error {
	err := ValidateCapabilityMatrix(a.config.Matrix)
	if err != nil {
		return err
	}

	a.agent.OnDegradationChanged(func(level fdm.DegradationLevel) {
		a.updateCapabilities(level)
		for _, cb := range a.cbs {
			cb(level)
		}
	})

	a.updateCapabilities(a.agent.LocalDegradationLevel())
	return nil
}

func (a *agentImpl) Stop() error {
	a.cancel()
	return nil
}

func (a *agentImpl) ClusterServiceState() ClusterServiceState {
	return a.currentState
}

func (a *agentImpl) IsCapable(cap Capability) bool {
	for _, c := range a.activeCaps {
		if c == cap {
			return true
		}
	}
	return false
}

func (a *agentImpl) RequireCapability(cap Capability) error {
	if !a.IsCapable(cap) {
		return ErrCapabilityUnavailable
	}
	return nil
}

func (a *agentImpl) OnDegradationChanged(callback func(level fdm.DegradationLevel)) {
	a.cbs = append(a.cbs, callback)
}

func (a *agentImpl) ActiveCapabilities() []Capability {
	return a.activeCaps
}

func (a *agentImpl) updateCapabilities(level fdm.DegradationLevel) {
	caps, ok := a.config.Matrix[level]
	if !ok {
		a.activeCaps = []Capability{}
		a.currentState = ServiceStateOffline
		return
	}
	a.activeCaps = caps
	if level == fdm.DegradationNone {
		a.currentState = ServiceStateHealthy
	} else if level == fdm.DegradationAll {
		a.currentState = ServiceStateOffline
	} else {
		a.currentState = ServiceStateDegraded
	}
}

//Personal.AI order the ending