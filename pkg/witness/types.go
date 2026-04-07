package witness

import "time"

type WitnessRole string

const (
	RoleArbiter WitnessRole = "arbiter"
	RoleVoter   WitnessRole = "voter"
)

type WitnessStatus string

const (
	StatusHealthy   WitnessStatus = "healthy"
	StatusUnhealthy WitnessStatus = "unhealthy"
	StatusUnknown   WitnessStatus = "unknown"
)

type WitnessConfig struct {
	Endpoints []string
	TimeoutMs int
}

type WitnessState struct {
	VMID      string
	Available bool
	CheckedAt time.Time
	Reason    string
}

type ConfirmationRequest struct {
	NodeID string
}

type ConfirmationResponse struct {
	Confirmed bool
	Error     error
}

