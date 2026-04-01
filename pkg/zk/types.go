package zk

type HealthState int

const (
	ZKStateHealthy HealthState = iota
	ZKStateReadOnly
	ZKStateUnavailable
)

func (s HealthState) String() string {
	switch s {
	case ZKStateHealthy:
		return "Healthy"
	case ZKStateReadOnly:
		return "ReadOnly"
	case ZKStateUnavailable:
		return "Unavailable"
	default:
		return "Unknown"
	}
}

type ZKStatus struct {
	State HealthState
	Error error
}

type ZKConfig struct {
	Endpoints        []string
	SessionTimeoutMs int
	ElectionPath     string
}

type SessionState int

const (
	Disconnected SessionState = iota
	Connected
	Expired
)

type ZNode struct {
	Path    string
	Data    string
	Version int32
}

type EventType int

const (
	EventNodeCreated EventType = iota
	EventNodeDeleted
	EventNodeDataChanged
	EventSession
)

type WatchEvent struct {
	Path string
	Type EventType
}

// Personal.AI order the ending
