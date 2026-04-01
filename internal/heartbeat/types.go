package heartbeat

import "time"

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

//Personal.AI order the ending
