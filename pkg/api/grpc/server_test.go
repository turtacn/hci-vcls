package grpc

import (
	"context"
	"testing"

	"github.com/turtacn/hci-vcls/internal/app"
	"github.com/turtacn/hci-vcls/pkg/api/proto"
)

import (
	"github.com/turtacn/hci-vcls/internal/config"
	"github.com/turtacn/hci-vcls/internal/election"
	"go.uber.org/zap"
)

func TestNewServer(t *testing.T) {
	cfg := &config.Config{}
	svc := app.NewService(cfg, zap.NewNop(), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	server := NewServer(svc)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestServer_GetVersion(t *testing.T) {
	cfg := &config.Config{}
	svc := app.NewService(cfg, zap.NewNop(), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	server := NewServer(svc)
	resp, err := server.GetVersion(context.Background(), &proto.VersionRequest{})
	if err != nil {
		t.Fatalf("GetVersion returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("GetVersion returned nil response")
	}
	if resp.Version != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", resp.Version)
	}
}

func TestServer_GetStatus(t *testing.T) {
	cfg := &config.Config{}
	elector := election.NewMemoryElector("n1", nil)
	svc := app.NewService(cfg, zap.NewNop(), nil, elector, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	server := NewServer(svc)

	resp, err := server.GetStatus(context.Background(), &proto.StatusRequest{})
	if err != nil {
		t.Fatalf("GetStatus returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("GetStatus returned nil response")
	}
}

func TestServer_GetDegradation(t *testing.T) {
	cfg := &config.Config{}
	elector := election.NewMemoryElector("n1", nil)
	svc := app.NewService(cfg, zap.NewNop(), nil, elector, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	server := NewServer(svc)

	resp, err := server.GetDegradation(context.Background(), &proto.DegradationRequest{})
	if err != nil {
		t.Fatalf("GetDegradation returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("GetDegradation returned nil response")
	}
}

func TestServer_EvaluateHA(t *testing.T) {
	cfg := &config.Config{}
	svc := app.NewService(cfg, zap.NewNop(), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	server := NewServer(svc)
	// it will fail with not leader error
	_, err := server.EvaluateHA(context.Background(), &proto.EvaluateHARequest{})
	if err == nil {
		t.Fatalf("Expected err returned nil response")
	}
}

func TestServer_ListTasks(t *testing.T) {
	cfg := &config.Config{}
	svc := app.NewService(cfg, zap.NewNop(), nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	server := NewServer(svc)
	resp, err := server.ListTasks(context.Background(), &proto.ListTasksRequest{})
	if err != nil {
		t.Fatalf("ListTasks returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("ListTasks returned nil response")
	}
}

func TestServer_GetPlan(t *testing.T) {
	server := &Server{}
	resp, err := server.GetPlan(context.Background(), &proto.GetPlanRequest{PlanId: "123"})
	if err != nil {
		t.Fatalf("GetPlan returned error: %v", err)
	}
	if resp.PlanId != "123" {
		t.Errorf("expected plan id 123, got %s", resp.PlanId)
	}
}

func TestServer_SweeperStatus(t *testing.T) {
	server := &Server{}
	resp, err := server.SweeperStatus(context.Background(), &proto.SweeperStatusRequest{})
	if err != nil {
		t.Fatalf("SweeperStatus returned error: %v", err)
	}
	if resp.ReleasedCount != 0 {
		t.Errorf("expected released count 0, got %d", resp.ReleasedCount)
	}
}

func TestServer_QueryAudit(t *testing.T) {
	server := &Server{}
	resp, err := server.QueryAudit(context.Background(), &proto.QueryAuditRequest{})
	if err != nil {
		t.Fatalf("QueryAudit returned error: %v", err)
	}
	if len(resp.Records) != 0 {
		t.Errorf("expected 0 records, got %d", len(resp.Records))
	}
}
