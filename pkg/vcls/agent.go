package vcls

import (
	"context"

	"github.com/turtacn/hci-vcls/pkg/fdm"
)

type Agent interface {
	Start(ctx context.Context) error
	Stop() error
	ClusterServiceState() ClusterServiceState
	IsCapable(cap Capability) bool
	RequireCapability(cap Capability) error
	OnDegradationChanged(callback func(level fdm.DegradationLevel))
	ActiveCapabilities() []Capability
}

// Personal.AI order the ending
