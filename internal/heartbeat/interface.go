package heartbeat

import "time"

type Monitor interface {
	Record(sample Sample)
	GetSummary(nodeID string) (Summary, bool)
	ListSummaries(clusterID string) []Summary
	CheckTimeouts(now time.Time, timeout time.Duration)
}

