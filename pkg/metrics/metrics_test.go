package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNoopMetrics(t *testing.T) {
	m := NewNoopMetrics()

	// These should not panic
	m.IncElectionTotal("node-1", "success")
	m.IncLeaderChange("cluster-1")
	m.IncHeartbeatLost("node-1", "cluster-1")
	m.SetDegradationLevel("cluster-1", 1.0)
	m.IncHATaskTotal("cluster-1", "completed")
	m.ObserveHAExecutionDuration("cluster-1", 2.5)
	m.SetProtectedVMCount("cluster-1", 10)

	// Add these explicitly to cover noop methods completely
	m.IncSweeperReleaseOK()
	m.IncSweeperReleaseFailed()
	m.SetSweeperLastRunUnix(float64(1000))
	m.IncStateMachineTransition("stable", "degraded", "heartbeat_lost")
	m.SetStateMachineCurrentState("degraded")
	m.ObserveEvaluationDuration(1.2)
}

func TestPrometheusMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, err := NewPrometheusMetrics(reg)
	if err != nil {
		t.Fatalf("Failed to create PrometheusMetrics: %v", err)
	}

	m.IncElectionTotal("node-1", "success")
	m.IncLeaderChange("cluster-1")
	m.IncHeartbeatLost("node-1", "cluster-1")
	m.SetDegradationLevel("cluster-1", 1.0)
	m.IncHATaskTotal("cluster-1", "completed")
	m.ObserveHAExecutionDuration("cluster-1", 2.5)
	m.SetProtectedVMCount("cluster-1", 10)

	m.IncSweeperReleaseOK()
	m.IncSweeperReleaseFailed()
	m.SetSweeperLastRunUnix(float64(1000))
	m.IncStateMachineTransition("stable", "degraded", "heartbeat_lost")
	m.SetStateMachineCurrentState("degraded")
	m.ObserveEvaluationDuration(1.2)

	_, err = reg.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}
}

