package app

import (
	"context"

	"github.com/turtacn/hci-vcls/internal/app/audit"
)

type haAuditAdapter struct {
	logger audit.AuditLogger
}

func NewHAAuditAdapter(l audit.AuditLogger) *haAuditAdapter {
	return &haAuditAdapter{logger: l}
}

func (a *haAuditAdapter) LogHADecision(ctx context.Context, clusterID, vmid, planID, bootPath, sourceHost, targetHost, reason, degradation, outcome, errStr string) error {
	return a.logger.LogHADecision(ctx, audit.HADecisionRecord{
		ClusterID:   clusterID,
		VMID:        vmid,
		PlanID:      planID,
		BootPath:    bootPath,
		SourceHost:  sourceHost,
		TargetHost:  targetHost,
		Reason:      reason,
		Degradation: degradation,
		Outcome:     outcome,
		Error:       errStr,
	})
}
