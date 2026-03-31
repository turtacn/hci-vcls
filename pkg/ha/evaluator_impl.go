package ha

import (
	"context"

	"github.com/turtacn/hci-vcls/pkg/cache"
	"github.com/turtacn/hci-vcls/pkg/fdm"
)

type evaluatorImpl struct {
	cacheManager cache.CacheManager
	fdmAgent     fdm.Agent
}

func NewEvaluator(cacheManager cache.CacheManager, fdmAgent fdm.Agent) Evaluator {
	return &evaluatorImpl{
		cacheManager: cacheManager,
		fdmAgent:     fdmAgent,
	}
}

func (e *evaluatorImpl) Evaluate(ctx context.Context, vmid string, cv ClusterView) (HADecision, error) {
	path, err := e.SelectBootPath(ctx, vmid)
	if err != nil {
		return HADecision{Action: ActionSkip, Reason: err.Error()}, err
	}

	targetNode, err := e.SelectTargetNode(ctx, vmid, cv)
	if err != nil {
		return HADecision{Action: ActionSkip, Reason: err.Error()}, err
	}

	priority, err := e.AssignPriority(ctx, vmid)
	if err != nil {
		return HADecision{Action: ActionSkip, Reason: err.Error()}, err
	}

	return HADecision{
		VMID:       vmid,
		Action:     ActionBoot,
		Path:       path,
		TargetNode: targetNode,
		Priority:   priority,
		Reason:     "Ready for HA Boot",
	}, nil
}

func (e *evaluatorImpl) SelectBootPath(ctx context.Context, vmid string) (BootPath, error) {
	// Simple strategy: check ZK health and MySQL health
	// Here, we just return a default path for the mock implementation
	return BootPathMySQL, nil
}

func (e *evaluatorImpl) SelectTargetNode(ctx context.Context, vmid string, cv ClusterView) (string, error) {
	// A real implementation would filter out dead nodes or the node the VM was originally on.
	// We select the first available node from the cluster view.
	for nodeID, state := range cv.Nodes {
		if state == fdm.NodeStateAlive && nodeID != e.fdmAgent.LeaderNodeID() {
			return nodeID, nil
		}
	}
	return "", ErrInsufficientResources
}

func (e *evaluatorImpl) AssignPriority(ctx context.Context, vmid string) (int, error) {
	// Mock priority
	return 1, nil
}

//Personal.AI order the ending