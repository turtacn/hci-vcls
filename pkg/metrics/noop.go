package metrics

type NoopMetrics struct{}

func NewNoopMetrics() Metrics {
	return &NoopMetrics{}
}

func (m *NoopMetrics) IncHABootTotal(labels MetricLabels)                                 {}
func (m *NoopMetrics) ObserveHABootDuration(durationSeconds float64, labels MetricLabels) {}
func (m *NoopMetrics) IncFDMHeartbeatLost(labels MetricLabels)                            {}
func (m *NoopMetrics) SetDegradationLevel(level float64, labels MetricLabels)             {}
func (m *NoopMetrics) SetCacheAgeSeconds(ageSeconds float64, labels MetricLabels)         {}
func (m *NoopMetrics) IncElectionTotal(labels MetricLabels)                               {}
func (m *NoopMetrics) IncLeaderChanges(labels MetricLabels)                               {}

//Personal.AI order the ending
