package statemachine

import (
	"time"

	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/zk"
)

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

type StateMachineConfig struct {
	EvaluationIntervalMs int
	CooldownMs           int
}

//Personal.AI order the ending