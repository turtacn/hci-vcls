package grpc

import (
	"context"

	"github.com/turtacn/hci-vcls/internal/app"
	"github.com/turtacn/hci-vcls/pkg/api/proto"
)

type Server struct {
	proto.UnimplementedHCIVclsServiceServer

	svc *app.Service
}

func NewServer(svc *app.Service) *Server {
	return &Server{
		svc: svc,
	}
}

func (s *Server) GetVersion(ctx context.Context, req *proto.VersionRequest) (*proto.VersionResponse, error) {
	return &proto.VersionResponse{
		Version: "1.0.0",
		Commit:  "unknown",
		Date:    "unknown",
	}, nil
}

func (s *Server) GetStatus(ctx context.Context, req *proto.StatusRequest) (*proto.StatusResponse, error) {
	status := s.svc.Status()
	return &proto.StatusResponse{
		IsLeader:         status.IsLeader,
		LeaderId:         status.Leader,
		ClusterState:     status.ClusterState,
		DegradationLevel: status.DegradationLevel,
	}, nil
}

func (s *Server) GetDegradation(ctx context.Context, req *proto.DegradationRequest) (*proto.DegradationResponse, error) {
	status := s.svc.Status()
	return &proto.DegradationResponse{
		Level: status.DegradationLevel,
	}, nil
}

func (s *Server) EvaluateHA(ctx context.Context, req *proto.EvaluateHARequest) (*proto.EvaluateHAResponse, error) {
	plan, err := s.svc.EvaluateHA(ctx, req.ClusterId)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return &proto.EvaluateHAResponse{PlanId: ""}, nil
	}
	return &proto.EvaluateHAResponse{
		PlanId: plan.ID,
	}, nil
}

func (s *Server) ListTasks(ctx context.Context, req *proto.ListTasksRequest) (*proto.ListTasksResponse, error) {
	return &proto.ListTasksResponse{
		Tasks: make([]*proto.ListTasksResponse_TaskInfo, 0),
	}, nil
}

