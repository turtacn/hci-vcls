package fdm

import "time"

type DegradationLevel string

const (
	DegradationNone     DegradationLevel = "None"
	DegradationMinor    DegradationLevel = "Minor"
	DegradationMajor    DegradationLevel = "Major"
	DegradationCritical DegradationLevel = "Critical"
)

type HostState struct {
	NodeID        string
	ClusterID     string
	Healthy       bool
	LostCount     int
	LastHeartbeat time.Time
	FaultDomain   string
}

type ClusterState struct {
	ClusterID        string
	Hosts            []HostState
	Degradation      DegradationLevel
	HeartbeatLossSum int
	UnhealthyHosts   []string
	QuorumRisk       bool
	LeaderAtRisk     bool
	Reason           string
	ComputedAt       time.Time
}

// Below are older types left for backward compatibility to keep compilation alive
// in case other packages are relying on them temporarily.
type OldDegradationLevel int

const (
	OldDegradationNone OldDegradationLevel = iota
	OldDegradationZK
	OldDegradationCFS
	OldDegradationMySQL
	OldDegradationAll
)

type NodeState string

const (
	NodeStateAlive   NodeState = "alive"
	NodeStateDead    NodeState = "dead"
	NodeStateUnknown NodeState = "unknown"
)

type HeartbeatLevel int

const (
	HeartbeatL0 HeartbeatLevel = iota
	HeartbeatL1
	HeartbeatL2
)

type FDMConfig struct {
	NodeID          string
	ClusterID       string
	HeartbeatPeers  []string
	ProbeIntervalMs int
}

type ClusterView struct {
	LeaderID         string
	Nodes            map[string]NodeState
	DegradationLevel OldDegradationLevel
}

