package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func TestNoopMetrics(t *testing.T) {
	m := NewNoopMetrics()
	labels := MetricLabels{LabelNodeID: "test-node"}

	// These should not panic
	m.IncHABootTotal(labels)
	m.ObserveHABootDuration(1.0, labels)
	m.IncFDMHeartbeatLost(labels)
	m.SetDegradationLevel(0, labels)
	m.SetCacheAgeSeconds(10, labels)
	m.IncElectionTotal(labels)
	m.IncLeaderChanges(labels)
}

func TestPrometheusMetrics(t *testing.T) {
	reg := prometheus.NewRegistry()
	m, err := NewPrometheusMetrics(reg)
	if err != nil {
		t.Fatalf("Failed to create PrometheusMetrics: %v", err)
	}

	labels := MetricLabels{
		LabelClusterID: "c1",
		LabelNodeID:    "n1",
		LabelVMID:      "v1",
		LabelBootPath:  "mysql",
		LabelResult:    "success",
	}

	m.IncHABootTotal(labels)
	m.ObserveHABootDuration(2.5, labels)

	simpleLabels := MetricLabels{
		LabelClusterID: "c1",
		LabelNodeID:    "n1",
	}
	m.IncFDMHeartbeatLost(simpleLabels)
	m.IncElectionTotal(simpleLabels)
	m.IncLeaderChanges(simpleLabels)
	m.SetCacheAgeSeconds(5.0, simpleLabels)

	degLabels := MetricLabels{
		LabelClusterID:        "c1",
		LabelDegradationLevel: "0",
	}
	m.SetDegradationLevel(0, degLabels)

	// Since we are just testing the wiring, gathering the registry is enough
	// to ensure they are registered without errors.
	_, err = reg.Gather()
	if err != nil {
		t.Fatalf("Failed to gather metrics: %v", err)
	}
}

//Personal.AI order the ending