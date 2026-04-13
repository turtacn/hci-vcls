package statemachine

import (
	"testing"

	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/zk"
)

type mockMetrics struct{}

func (m *mockMetrics) IncElectionTotal(node, result string)                       {}
func (m *mockMetrics) IncLeaderChange(cluster string)                             {}
func (m *mockMetrics) IncHeartbeatLost(node, cluster string)                      {}
func (m *mockMetrics) SetDegradationLevel(cluster string, level float64)          {}
func (m *mockMetrics) IncHATaskTotal(cluster, status string)                      {}
func (m *mockMetrics) ObserveHAExecutionDuration(cluster string, seconds float64) {}
func (m *mockMetrics) SetProtectedVMCount(cluster string, count float64)          {}
func (m *mockMetrics) IncSweeperReleaseOK()                                       {}
func (m *mockMetrics) IncSweeperReleaseFailed()                                   {}
func (m *mockMetrics) SetSweeperLastRunUnix(ts float64)                           {}
func (m *mockMetrics) IncStateMachineTransition(from, to, event string)           {}
func (m *mockMetrics) SetStateMachineCurrentState(state string)                   {}
func (m *mockMetrics) ObserveEvaluationDuration(seconds float64)                  {}

func TestMachine(t *testing.T) {
	m := NewMachine(&mockMetrics{})

	if m.Current() != StateInit {
		t.Errorf("Expected initial state StateInit, got %v", m.Current())
	}

	// Test invalid transition
	if m.CanTransition(EventFailoverCompleted) {
		t.Errorf("Should not be able to transition EventFailoverCompleted from StateInit")
	}
	err := m.Transition(EventFailoverCompleted)
	if err != ErrIllegalTransition {
		t.Errorf("Expected ErrIllegalTransition, got %v", err)
	}

	// Test valid transitions
	events := []Event{
		EventHeartbeatRestored,   // Init -> Stable
		EventDegradationDetected, // Stable -> Degraded
		EventEvaluationStarted,   // Degraded -> Evaluating
		EventFailoverTriggered,   // Evaluating -> Failover
		EventFailoverCompleted,   // Failover -> Recovered
		EventHeartbeatRestored,   // Recovered -> Stable
	}

	for _, e := range events {
		if !m.CanTransition(e) {
			t.Errorf("Expected CanTransition %v from %v to be true", e, m.Current())
		}
		err := m.Transition(e)
		if err != nil {
			t.Errorf("Failed to transition %v: %v", e, err)
		}
	}

	if m.Current() != StateStable {
		t.Errorf("Expected final state StateStable, got %v", m.Current())
	}

	history := m.History()
	if len(history) != len(events) {
		t.Errorf("Expected history length %d, got %d", len(events), len(history))
	}
}

func TestEvaluate(t *testing.T) {
	tests := []struct {
		name     string
		input    EvaluationInput
		expected EvaluationResult
	}{
		{
			name: "FDM Critical",
			input: EvaluationInput{
				FDMLevel: fdm.DegradationCritical,
			},
			expected: EvaluationResult{Level: fdm.DegradationCritical, Reason: "FDM critical (Isolated)"},
		},
		{
			name: "ZK and CFS Read-Only",
			input: EvaluationInput{
				ZKStatus:  zk.ZKStatus{State: zk.ZKStateReadOnly},
				CFSStatus: cfs.CFSStatus{State: cfs.CFSStateReadOnly},
			},
			expected: EvaluationResult{Level: fdm.DegradationMajor, Reason: "ZK and CFS are read-only"},
		},
		{
			name: "MySQL Unavailable",
			input: EvaluationInput{
				MySQLStatus: mysql.MySQLStatus{State: mysql.MySQLStateUnavailable},
			},
			expected: EvaluationResult{Level: fdm.DegradationMajor, Reason: "MySQL unavailable"},
		},
		{
			name: "ZK Read-Only, MySQL OK",
			input: EvaluationInput{
				ZKStatus:    zk.ZKStatus{State: zk.ZKStateReadOnly},
				MySQLStatus: mysql.MySQLStatus{State: mysql.MySQLStateHealthy},
			},
			expected: EvaluationResult{Level: fdm.DegradationMinor, Reason: "ZK is read-only, MySQL is OK"},
		},
		{
			name: "Normal",
			input: EvaluationInput{
				ZKStatus:    zk.ZKStatus{State: zk.ZKStateHealthy},
				CFSStatus:   cfs.CFSStatus{State: cfs.CFSStateHealthy},
				MySQLStatus: mysql.MySQLStatus{State: mysql.MySQLStateHealthy},
			},
			expected: EvaluationResult{Level: fdm.DegradationNone, Reason: "Normal"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := Evaluate(tt.input)
			if res.Level != tt.expected.Level || res.Reason != tt.expected.Reason {
				t.Errorf("Evaluate() = %v, expected %v", res, tt.expected)
			}
		})
	}
}
