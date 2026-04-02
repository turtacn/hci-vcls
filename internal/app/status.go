package app

import "context"

type StatusResponse struct {
	Leader           string
	IsLeader         bool
	ClusterState     string
	DegradationLevel string
	ProtectedVMCount int
	PendingTasks     int
}

func (s *Service) Status() StatusResponse {
	resp := StatusResponse{
		IsLeader: s.election.IsLeader(),
		Leader:   s.election.Status().LeaderID,
	}

	if s.fdmAgent != nil {
		resp.DegradationLevel = string(s.fdmAgent.LocalDegradationLevel())
	}

	if s.statemachine != nil {
		resp.ClusterState = string(s.statemachine.Current())
	}

	if s.vcls != nil && s.config != nil {
		vms, _ := s.vcls.ListProtected(context.Background(), s.config.Node.ClusterID)
		resp.ProtectedVMCount = len(vms)
	}

	return resp
}

// Personal.AI order the ending
