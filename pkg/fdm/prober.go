package fdm

import "context"

type ProbeResult struct {
	Level   HeartbeatLevel
	Success bool
	Error   error
}

type Prober interface {
	ProbeL0(ctx context.Context) ProbeResult
	ProbeL1(ctx context.Context) ProbeResult
	ProbeL2(ctx context.Context) ProbeResult
	ProbeAll(ctx context.Context) map[HeartbeatLevel]ProbeResult
}

//Personal.AI order the ending
