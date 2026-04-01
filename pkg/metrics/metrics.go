package metrics

type Metrics interface {
	IncHABootTotal(labels MetricLabels)
	ObserveHABootDuration(durationSeconds float64, labels MetricLabels)
	IncFDMHeartbeatLost(labels MetricLabels)
	SetDegradationLevel(level float64, labels MetricLabels)
	SetCacheAgeSeconds(ageSeconds float64, labels MetricLabels)
	IncElectionTotal(labels MetricLabels)
	IncLeaderChanges(labels MetricLabels)
}

//Personal.AI order the ending
