package metrics

type NoopMetrics struct{}

var _ Metrics = &NoopMetrics{}

func NewNoopMetrics() *NoopMetrics {
	return &NoopMetrics{}
}

func (m *NoopMetrics) IncElectionTotal(node, result string)                       {}
func (m *NoopMetrics) IncLeaderChange(cluster string)                             {}
func (m *NoopMetrics) IncHeartbeatLost(node, cluster string)                      {}
func (m *NoopMetrics) SetDegradationLevel(cluster string, level float64)          {}
func (m *NoopMetrics) IncHATaskTotal(cluster, status string)                      {}
func (m *NoopMetrics) ObserveHAExecutionDuration(cluster string, seconds float64) {}
func (m *NoopMetrics) SetProtectedVMCount(cluster string, count float64)          {}

func (m *NoopMetrics) IncSweeperReleaseOK()             {}
func (m *NoopMetrics) IncSweeperReleaseFailed()         {}
func (m *NoopMetrics) SetSweeperLastRunUnix(ts float64) {}

func (m *NoopMetrics) IncStateMachineTransition(from, to, event string) {}
func (m *NoopMetrics) SetStateMachineCurrentState(state string)         {}
func (m *NoopMetrics) ObserveEvaluationDuration(seconds float64)        {}
