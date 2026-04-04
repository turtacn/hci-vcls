package vcls

import (
	"time"

	"github.com/turtacn/hci-vcls/pkg/fdm"
)

type Capability string

const (
	CapabilityHA             Capability = "HA"
	CapabilityDRS            Capability = "DRS"
	CapabilityFT             Capability = "FT"
	CapabilityVMMigration    Capability = "VM_MIGRATION"
	CapabilityStoragevMotion Capability = "STORAGE_VMOTION"
	CapabilitySnapshots      Capability = "SNAPSHOTS"
)

type ClusterServiceState string

const (
	ServiceStateHealthy  ClusterServiceState = "HEALTHY"
	ServiceStateDegraded ClusterServiceState = "DEGRADED"
	ServiceStateOffline  ClusterServiceState = "OFFLINE"
)

type CapabilityMatrix map[fdm.DegradationLevel][]Capability

type VCLSConfig struct {
	Matrix CapabilityMatrix
}

type PowerStatus string

const (
	PowerRunning PowerStatus = "running"
	PowerStopped PowerStatus = "stopped"
	PowerUnknown PowerStatus = "unknown"
)

type VM struct {
	ID               string
	ClusterID        string
	CurrentHost      string
	PowerState       PowerStatus
	Protected        bool
	WitnessAvailable bool
	HostHealthy      bool
	EligibleForHA    bool
	LastSyncAt       time.Time
}

type Status struct {
	ClusterID      string
	VMCount        int
	ProtectedCount int
	EligibleCount  int
	LastRefreshAt  time.Time
}

// Personal.AI order the ending
