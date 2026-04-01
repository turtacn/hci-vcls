package vcls

import "github.com/turtacn/hci-vcls/pkg/fdm"

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

//Personal.AI order the ending
