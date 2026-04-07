package election

import (
	"time"
)

type LeaderStatus struct {
	LeaderID  string
	IsLeader  bool
	Term      int64
	UpdatedAt time.Time
}

type ElectionConfig struct {
	NodeID           string
	ZKElectionPath   string
	SessionTimeoutMs int
	RetryIntervalMs  int
}

type LeaderInfo struct {
	NodeID string
	Term   uint64
}

