package grpc

import (
	"context"

	"github.com/turtacn/hci-vcls/pkg/api/proto"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/vcls"
)

type Server struct {
	proto.UnimplementedHAServiceServer
	proto.UnimplementedFDMServiceServer
	proto.UnimplementedStatusServiceServer

	haEngine  ha.HAEngine
	fdmAgent  fdm.Agent
	vclsAgent vcls.Agent
}

func NewServer(haEngine ha.HAEngine, fdmAgent fdm.Agent, vclsAgent vcls.Agent) *Server {
	return &Server{
		haEngine:  haEngine,
		fdmAgent:  fdmAgent,
		vclsAgent: vclsAgent,
	}
}

func (s *Server) Evaluate(ctx context.Context, req *proto.EvaluateRequest) (*proto.EvaluateResponse, error) {
	decision, err := s.haEngine.Evaluate(ctx, req.Vmid)
	if err != nil {
		return nil, err
	}
	return &proto.EvaluateResponse{
		Vmid:       decision.VMID,
		Action:     string(decision.Action),
		TargetNode: decision.TargetNode,
		Reason:     decision.Reason,
	}, nil
}

func (s *Server) GetActiveTasks(ctx context.Context, req *proto.GetTasksRequest) (*proto.GetTasksResponse, error) {
	tasks := s.haEngine.ActiveTasks()
	resp := &proto.GetTasksResponse{}
	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, &proto.TaskInfo{
			Vmid:   t.VMID,
			Status: string(t.Status),
		})
	}
	return resp, nil
}

func (s *Server) GetClusterStatus(ctx context.Context, req *proto.GetClusterStatusRequest) (*proto.GetClusterStatusResponse, error) {
	cv := s.fdmAgent.ClusterView()
	resp := &proto.GetClusterStatusResponse{
		LeaderId:   cv.LeaderID,
		NodeStates: make(map[string]string),
	}
	for id, state := range cv.Nodes {
		resp.NodeStates[id] = string(state)
	}
	return resp, nil
}

func (s *Server) GetDegradation(ctx context.Context, req *proto.GetDegradationRequest) (*proto.GetDegradationResponse, error) {
	return &proto.GetDegradationResponse{
		Level: int32(s.fdmAgent.LocalDegradationLevel()),
	}, nil
}

func (s *Server) GetFullStatus(ctx context.Context, req *proto.GetFullStatusRequest) (*proto.GetFullStatusResponse, error) {
	level := s.fdmAgent.LocalDegradationLevel()
	caps := s.vclsAgent.ActiveCapabilities()

	capStrings := make([]string, 0, len(caps))
	for _, c := range caps {
		capStrings = append(capStrings, string(c))
	}

	return &proto.GetFullStatusResponse{
		LeaderId:           s.fdmAgent.LeaderNodeID(),
		DegradationLevel:   int32(level),
		ActiveCapabilities: capStrings,
	}, nil
}

//Personal.AI order the ending
