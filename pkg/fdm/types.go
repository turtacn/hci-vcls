package fdm

type DegradationLevel int

const (
	DegradationNone DegradationLevel = iota
	DegradationZK
	DegradationCFS
	DegradationMySQL
	DegradationAll
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
	DegradationLevel DegradationLevel
}

//Personal.AI order the ending
