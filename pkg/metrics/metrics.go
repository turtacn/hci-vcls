package metrics

type Metrics interface {
	IncElectionTotal(node, result string)
	IncLeaderChange(cluster string)
	IncHeartbeatLost(node, cluster string)
	SetDegradationLevel(cluster string, level float64)
	IncHATaskTotal(cluster, status string)
	ObserveHAExecutionDuration(cluster string, seconds float64)
	SetProtectedVMCount(cluster string, count float64)
	IncSweeperReleaseOK()
	IncSweeperReleaseFailed()
	SetSweeperLastRunUnix(ts float64)
	IncStateMachineTransition(from, to, event string)
	SetStateMachineCurrentState(state string)
	ObserveEvaluationDuration(seconds float64)
}
