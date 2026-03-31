package statemachine

import (
	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/zk"
)

func Evaluate(input EvaluationInput) EvaluationResult {
	if input.FDMLevel == fdm.DegradationAll || input.ZKStatus.State == zk.ZKStateUnavailable {
		return EvaluationResult{Level: fdm.DegradationZK, Reason: "ZK Unavailable or FDM All Degradation"}
	}
	if input.CFSStatus.State == cfs.CFSStateUnavailable || input.CFSStatus.State == cfs.CFSStateUnmounted {
		return EvaluationResult{Level: fdm.DegradationCFS, Reason: "CFS Unavailable"}
	}
	if input.MySQLStatus.State == mysql.MySQLStateUnavailable {
		return EvaluationResult{Level: fdm.DegradationMySQL, Reason: "MySQL Unavailable"}
	}
	if input.ZKStatus.State == zk.ZKStateReadOnly || input.CFSStatus.State == cfs.CFSStateReadOnly || input.MySQLStatus.State == mysql.MySQLStateReadOnly {
		return EvaluationResult{Level: fdm.DegradationAll, Reason: "Read Only State Detected"}
	}
	return EvaluationResult{Level: fdm.DegradationNone, Reason: "Healthy"}
}

func IsValidTransition(from, to fdm.DegradationLevel) bool {
	// For simplicity, allow all transitions initially. This can be refined later.
	return true
}

//Personal.AI order the ending