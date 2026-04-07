package heartbeat

import "time"

type Sample struct {
	NodeID     string
	ClusterID  string
	ReceivedAt time.Time
}

type Summary struct {
	NodeID     string
	ClusterID  string
	LastSeenAt time.Time
	LostCount  int
	Healthy    bool
	ObservedAt time.Time
}

// Old types to keep compilation for other packages
type HeartbeatLevel int

const (
	LevelL0 HeartbeatLevel = iota
	LevelL1
	LevelL2
)

type HeartbeatPacket struct {
	NodeID    string
	Timestamp time.Time
	Level     HeartbeatLevel
}

type HeartbeatConfig struct {
	IntervalMs int
	TimeoutMs  int
	NodeID     string
	Peers      []string
}

type HeartbeatState struct {
	LastSeen time.Time
	IsAlive  bool
}

type StateDigest struct {
	NodeID      string    `json:"node_id"`
	Term        int64     `json:"term"`
	CandidateID string    `json:"candidate_id"`
	Timestamp   time.Time `json:"timestamp"`
	Level       int       `json:"level"`
}

