package statemachine

import (
	"time"

	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/zk"
)

type State string

const (
	StateInit       State = "init"
	StateStable     State = "stable"
	StateDegraded   State = "degraded"
	StateEvaluating State = "evaluating" // 新增：FDM评估中
	StateFailover   State = "failover"
	StateRecovered  State = "recovered"
)

type Event string

const (
	EventHeartbeatRestored   Event = "heartbeat_restored"
	EventHeartbeatLost       Event = "heartbeat_lost"
	EventDegradationDetected Event = "degradation_detected"
	EventEvaluationStarted   Event = "evaluation_started"
	EventFailoverTriggered   Event = "failover_triggered"
	EventFailoverCompleted   Event = "failover_completed"
	EventRecoveredSignal     Event = "recovered_signal"
)

type StateTransition struct {
	From      State
	To        State
	Event     Event
	Timestamp time.Time
}

type Transition struct {
	From      fdm.DegradationLevel
	To        fdm.DegradationLevel
	Timestamp time.Time
	Reason    string
}

type EvaluationInput struct {
	ZKStatus    zk.ZKStatus
	CFSStatus   cfs.CFSStatus
	MySQLStatus mysql.MySQLStatus
	FDMLevel    fdm.DegradationLevel
}

type EvaluationResult struct {
	Level  fdm.DegradationLevel
	Reason string
}

type Capability string

const (
	CapabilityNormalBoot   Capability = "NORMAL_BOOT"
	CapabilityMinorityBoot Capability = "MINORITY_BOOT"
	CapabilityCacheRead    Capability = "CACHE_READ"
	CapabilityNoBoot       Capability = "NO_BOOT"
)

type StateMachineConfig struct {
	EvaluationIntervalMs int
	CooldownMs           int
}

// Personal.AI order the ending
