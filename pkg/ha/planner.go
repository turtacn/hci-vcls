package ha

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/turtacn/hci-vcls/pkg/vcls"
)

type plannerImpl struct{}

var _ Planner = &plannerImpl{}

func NewPlanner() Planner {
	return &plannerImpl{}
}

func (p *plannerImpl) BuildPlan(ctx context.Context, req PlanRequest) (*Plan, error) {
	// Step 1: Validate input
	if req.ClusterID == "" {
		return nil, ErrEmptyClusterID
	}
	if len(req.ProtectedVMs) == 0 {
		return nil, ErrNoProtectedVMs
	}
	if len(req.HostCandidates) == 0 {
		return nil, ErrNoCandidateHost
	}

	failedHostSet := make(map[string]bool)
	for _, h := range req.FailedHosts {
		failedHostSet[h] = true
	}

	var eligibleVMs []*vcls.VM
	// Step 2: Filter VMs
	for _, vm := range req.ProtectedVMs {
		if vm.EligibleForHA && failedHostSet[vm.CurrentHost] {
			eligibleVMs = append(eligibleVMs, vm)
		}
	}

	if len(eligibleVMs) == 0 {
		return nil, ErrNoProtectedVMs
	}

	candidates := make([]HostCandidate, len(req.HostCandidates))
	copy(candidates, req.HostCandidates)

	var tasks []VMTask

	hostDomainMap := make(map[string]string)
	for _, c := range candidates {
		hostDomainMap[c.HostID] = c.FaultDomain
	}

	// Step 4: Choose best host per VM
	for _, vm := range eligibleVMs {
		sourceDomain := hostDomainMap[vm.CurrentHost]

		bestHostIndex := -1
		var bestScore float64 = -1
		var bestReason string

		for i, host := range candidates {
			if !host.Healthy {
				continue
			}
			if host.HostID == vm.CurrentHost {
				continue
			}

			score, reason := ScoreHost(host, sourceDomain, req.PreferWitness)
			if score > bestScore {
				bestScore = score
				bestReason = reason
				bestHostIndex = i
			}
		}

		task := VMTask{
			ID:         uuid.NewString(),
			VMID:       vm.ID,
			ClusterID:  vm.ClusterID,
			SourceHost: vm.CurrentHost,
			Status:     TaskPending,
			Score:      bestScore,
			Reason:     bestReason,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		if bestHostIndex == -1 {
			task.Status = TaskSkipped
			task.Reason = "no healthy candidate host available"
		} else {
			task.TargetHost = candidates[bestHostIndex].HostID
			// Step 5: Choose BootPath
			if req.PreferWitness && candidates[bestHostIndex].WitnessCapable {
				task.BootPath = BootPathWitness
			} else if req.DegradationLevel != "" && req.DegradationLevel != "None" {
				task.BootPath = BootPathMinority
			} else {
				task.BootPath = BootPathNormal
			}

			// increment load
			candidates[bestHostIndex].CurrentLoad++
		}

		tasks = append(tasks, task)
	}

	// Step 6: Batching
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 5 // default
	}

	var finalTasks []VMTask
	batchNo := 1
	countInBatch := 0
	for _, t := range tasks {
		if t.Status != TaskSkipped {
			t.BatchNo = batchNo
			countInBatch++
			if countInBatch == batchSize {
				batchNo++
				countInBatch = 0
			}
		} else {
			t.BatchNo = 0
		}
		finalTasks = append(finalTasks, t)
	}

	totalBatches := batchNo
	if countInBatch == 0 && batchNo > 1 {
		totalBatches = batchNo - 1
	}

	// Step 7: Generate Plan
	return &Plan{
		ID:           uuid.NewString(),
		ClusterID:    req.ClusterID,
		Trigger:      "auto",
		Degradation:  req.DegradationLevel,
		Tasks:        finalTasks,
		TotalBatches: totalBatches,
		CreatedAt:    time.Now(),
	}, nil
}

// ScoreHost is exposed for testing
func ScoreHost(host HostCandidate, sourceFaultDomain string, preferWitness bool) (float64, string) {
	if !host.Healthy {
		return 0, fmt.Sprintf("HostID=%s: unhealthy", host.HostID)
	}

	var score float64 = 100
	var reasons []string
	reasons = append(reasons, "healthy(+100)")

	if host.CurrentLoad == 0 {
		score += 20
		reasons = append(reasons, "low-load(+20)")
	} else if host.CurrentLoad <= 3 {
		score += 10
		reasons = append(reasons, "low-load(+10)")
	}

	if host.FaultDomain != "" && host.FaultDomain != sourceFaultDomain {
		score += 15
		reasons = append(reasons, "cross-domain(+15)")
	}

	if host.RecentFailures == 0 {
		score += 10
		reasons = append(reasons, "no-recent-failures(+10)")
	} else if host.RecentFailures <= 2 {
		score += 5
		reasons = append(reasons, "few-recent-failures(+5)")
	}

	if preferWitness && host.WitnessCapable {
		score += 30
		reasons = append(reasons, "witness(+30)")
	}

	penalty := float64(host.CurrentLoad * 2)
	if penalty > 0 {
		score -= penalty
		reasons = append(reasons, fmt.Sprintf("load-penalty(-%d)", int(penalty)))
	}

	return score, fmt.Sprintf("HostID=%s: %s", host.HostID, strings.Join(reasons, ", "))
}

