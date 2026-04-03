package fdm

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type evaluatorImpl struct{}

var _ Evaluator = &evaluatorImpl{}

func NewEvaluator() Evaluator {
	return &evaluatorImpl{}
}

func (e *evaluatorImpl) Evaluate(ctx context.Context, clusterID string, leaderNodeID string, hosts []HostState, witnessAvailable bool) (*ClusterState, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	state := &ClusterState{
		ClusterID:        clusterID,
		Hosts:            hosts,
		Degradation:      DegradationNone,
		ComputedAt:       time.Now(),
		UnhealthyHosts:   []string{},
		HeartbeatLossSum: 0,
		Reason:           "",
	}

	if len(hosts) == 0 {
		state.Reason = "no hosts"
		return state, nil
	}

	healthyCount := 0
	unhealthyCount := 0
	total := len(hosts)

	for _, host := range hosts {
		if host.Healthy {
			healthyCount++
		} else {
			unhealthyCount++
			state.UnhealthyHosts = append(state.UnhealthyHosts, host.NodeID)
			if host.NodeID == leaderNodeID {
				state.LeaderAtRisk = true
			}
		}
		state.HeartbeatLossSum += host.LostCount
	}

	if unhealthyCount == 0 {
		state.Reason = "all hosts healthy"
		return state, nil
	}

	var reasons []string

	// v1 rule
	var v1Level DegradationLevel
	switch {
	case unhealthyCount == 1:
		v1Level = DegradationMinor
	case unhealthyCount == 2:
		v1Level = DegradationMajor
	case unhealthyCount >= 3:
		v1Level = DegradationCritical
	}

	// v2 rule
	var v2Level DegradationLevel
	ratio := float64(unhealthyCount) / float64(total)
	switch {
	case ratio <= 0.25:
		v2Level = DegradationMinor
	case ratio <= 0.50:
		v2Level = DegradationMajor
	default:
		v2Level = DegradationCritical
	}

	// Max of v1 and v2
	state.Degradation = maxLevel(v1Level, v2Level)
	reasons = append(reasons, fmt.Sprintf("unhealthy count: %d, ratio: %.2f", unhealthyCount, ratio))

	// v3 rules (Quorum, Leader, Witness)
	quorum := total/2 + 1
	if healthyCount < quorum {
		state.QuorumRisk = true
		state.Degradation = DegradationCritical
		reasons = append(reasons, "healthy count < quorum")
	}

	if state.LeaderAtRisk {
		state.Degradation = DegradationCritical
		reasons = append(reasons, "leader host unhealthy")
	}

	if !witnessAvailable && unhealthyCount >= 1 {
		// Minimum is Major
		if state.Degradation == DegradationNone || state.Degradation == DegradationMinor {
			state.Degradation = DegradationMajor
		}
		reasons = append(reasons, "witness unavailable and unhealthy >= 1")
	}

	state.Reason = strings.Join(reasons, "; ")
	return state, nil
}

func maxLevel(a, b DegradationLevel) DegradationLevel {
	if LevelWeight(a) >= LevelWeight(b) {
		return a
	}
	return b
}

// Personal.AI order the ending
