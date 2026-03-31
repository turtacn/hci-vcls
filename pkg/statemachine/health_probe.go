package statemachine

import (
	"context"

	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/zk"
)

type HealthProber struct {
	zkAdapter    zk.Adapter
	cfsAdapter   cfs.Adapter
	mysqlAdapter mysql.Adapter
	fdmAgent     fdm.Agent
}

func NewHealthProber(zk zk.Adapter, cfs cfs.Adapter, mysql mysql.Adapter, fdm fdm.Agent) *HealthProber {
	return &HealthProber{
		zkAdapter:    zk,
		cfsAdapter:   cfs,
		mysqlAdapter: mysql,
		fdmAgent:     fdm,
	}
}

func (p *HealthProber) Sample(ctx context.Context) EvaluationInput {
	var zStatus zk.ZKStatus
	if p.zkAdapter != nil {
		zStatus = p.zkAdapter.Health()
	} else {
		zStatus = zk.ZKStatus{State: zk.ZKStateHealthy}
	}

	var cStatus cfs.CFSStatus
	if p.cfsAdapter != nil {
		cStatus = p.cfsAdapter.Health()
	} else {
		cStatus = cfs.CFSStatus{State: cfs.CFSStateHealthy}
	}

	var mStatus mysql.MySQLStatus
	if p.mysqlAdapter != nil {
		mStatus = p.mysqlAdapter.Health()
	} else {
		mStatus = mysql.MySQLStatus{State: mysql.MySQLStateHealthy}
	}

	var fLevel fdm.DegradationLevel
	if p.fdmAgent != nil {
		fLevel = p.fdmAgent.LocalDegradationLevel()
	} else {
		fLevel = fdm.DegradationNone
	}

	return EvaluationInput{
		ZKStatus:    zStatus,
		CFSStatus:   cStatus,
		MySQLStatus: mStatus,
		FDMLevel:    fLevel,
	}
}

//Personal.AI order the ending