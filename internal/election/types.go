package election

type ElectionConfig struct {
	NodeID         string
	ZKElectionPath string
	SessionTimeoutMs int
	RetryIntervalMs  int
}

type LeaderInfo struct {
	NodeID string
	Term   uint64
}

//Personal.AI order the ending