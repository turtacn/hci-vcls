package statemachine

import (
	"errors"
	"sync"
	"time"

	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/zk"
)

var (
	ErrIllegalTransition = errors.New("statemachine: illegal transition")
)

type machineImpl struct {
	mu           sync.RWMutex
	current      State
	history      []StateTransition
	currentLevel string
	metrics      metrics.Metrics
}

var _ Machine = &machineImpl{}

func NewMachine(m metrics.Metrics) Machine {
	return &machineImpl{
		current:      StateInit,
		history:      make([]StateTransition, 0),
		currentLevel: string(fdm.DegradationNone),
		metrics:      m,
	}
}

// Transition(event string) logic wrapper
func (m *machineImpl) TransitionString(event string) error {
	return m.Transition(Event(event))
}

func (m *machineImpl) EvaluateWithInput(input interface{}) (string, string) {
	start := time.Now()
	defer func() {
		if m.metrics != nil {
			m.metrics.ObserveEvaluationDuration(time.Since(start).Seconds())
		}
	}()

	if evalMap, ok := input.(map[string]interface{}); ok {
		zkState := zk.ZKStateUnavailable
		if s, ok := evalMap["ZKState"].(int); ok {
			zkState = zk.HealthState(s)
		}
		cfsState := cfs.CFSStateUnavailable
		if s, ok := evalMap["CFSState"].(int); ok {
			cfsState = cfs.HealthState(s)
		}
		mysqlState := mysql.MySQLStateUnavailable
		if s, ok := evalMap["MySQLState"].(int); ok {
			mysqlState = mysql.HealthState(s)
		}
		fdmLevel := fdm.DegradationCritical
		if l, ok := evalMap["FDMLevel"].(fdm.DegradationLevel); ok {
			fdmLevel = l
		}

		res := Evaluate(EvaluationInput{
			ZKStatus:    zk.ZKStatus{State: zkState},
			CFSStatus:   cfs.CFSStatus{State: cfsState},
			MySQLStatus: mysql.MySQLStatus{State: mysqlState},
			FDMLevel:    fdmLevel,
		})
		m.mu.Lock()
		m.currentLevel = string(res.Level)
		m.mu.Unlock()
		return string(res.Level), res.Reason
	}

	evalInput, ok := input.(EvaluationInput)
	if !ok {
		return string(fdm.DegradationCritical), "invalid input type"
	}
	res := Evaluate(evalInput)
	m.mu.Lock()
	m.currentLevel = string(res.Level)
	m.mu.Unlock()
	return string(res.Level), res.Reason
}

func (m *machineImpl) CurrentLevel() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentLevel
}

// 迁移矩阵（完整定义，非法事件返回 error，不 panic，不改变状态）：
func (m *machineImpl) getNextState(current State, event Event) (State, bool) {
	switch current {
	case StateInit:
		if event == EventHeartbeatRestored {
			return StateStable, true
		}
	case StateStable:
		if event == EventDegradationDetected || event == EventHeartbeatLost {
			return StateDegraded, true
		}
	case StateDegraded:
		if event == EventEvaluationStarted {
			return StateEvaluating, true
		}
		if event == EventHeartbeatRestored {
			return StateStable, true
		}
	case StateEvaluating:
		if event == EventFailoverTriggered {
			return StateFailover, true
		}
		if event == EventHeartbeatRestored {
			return StateStable, true
		}
	case StateFailover:
		if event == EventFailoverCompleted {
			return StateRecovered, true
		}
	case StateRecovered:
		if event == EventHeartbeatRestored {
			return StateStable, true
		}
	}
	return current, false
}

func (m *machineImpl) Transition(event Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	nextState, valid := m.getNextState(m.current, event)
	if !valid {
		return ErrIllegalTransition
	}

	prevState := m.current
	m.history = append(m.history, StateTransition{
		From:      m.current,
		To:        nextState,
		Event:     event,
		Timestamp: time.Now(),
	})
	m.current = nextState

	if m.metrics != nil {
		m.metrics.IncStateMachineTransition(string(prevState), string(nextState), string(event))
		m.metrics.SetStateMachineCurrentState(string(nextState))
	}

	return nil
}

func (m *machineImpl) Current() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

func (m *machineImpl) CanTransition(event Event) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, valid := m.getNextState(m.current, event)
	return valid
}

func (m *machineImpl) History() []StateTransition {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h := make([]StateTransition, len(m.history))
	copy(h, m.history)
	return h
}

// Evaluate determines the degradation level based on the current cluster status inputs.
// It is a pure function.
func Evaluate(input EvaluationInput) EvaluationResult {
	if input.FDMLevel == fdm.DegradationCritical {
		return EvaluationResult{Level: fdm.DegradationCritical, Reason: "FDM critical (Isolated)"}
	}

	zkRO := input.ZKStatus.State == zk.ZKStateReadOnly
	cfsRO := input.CFSStatus.State == cfs.CFSStateReadOnly
	mysqlUnavail := input.MySQLStatus.State == mysql.MySQLStateUnavailable

	if zkRO && cfsRO {
		return EvaluationResult{Level: fdm.DegradationMajor, Reason: "ZK and CFS are read-only"}
	}

	if mysqlUnavail {
		return EvaluationResult{Level: fdm.DegradationMajor, Reason: "MySQL unavailable"}
	}

	if zkRO && !mysqlUnavail {
		return EvaluationResult{Level: fdm.DegradationMinor, Reason: "ZK is read-only, MySQL is OK"}
	}

	return EvaluationResult{Level: fdm.DegradationNone, Reason: "Normal"}
}

// MapCapabilities maps a degradation level to a set of operational capabilities.
func MapCapabilities(level fdm.DegradationLevel) []Capability {
	switch level {
	case fdm.DegradationNone:
		return []Capability{CapabilityNormalBoot}
	case fdm.DegradationMinor:
		// ZK read-only, but CFS and MySQL are ok
		return []Capability{CapabilityMinorityBoot}
	case fdm.DegradationMajor:
		// ZK read-only, CFS read-only, MySQL might be ok
		return []Capability{CapabilityMinorityBoot, CapabilityCacheRead}
	case fdm.DegradationCritical:
		return []Capability{CapabilityNoBoot}
	default:
		return []Capability{CapabilityNoBoot}
	}
}
