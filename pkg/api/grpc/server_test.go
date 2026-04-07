package grpc

import (
	"context"
	"testing"

	"github.com/turtacn/hci-vcls/internal/app"
	"github.com/turtacn/hci-vcls/pkg/api/proto"
)

func TestNewServer(t *testing.T) {
	svc := &app.Service{}
	server := NewServer(svc)
	if server == nil {
		t.Fatal("NewServer returned nil")
	}
}

func TestServer_GetVersion(t *testing.T) {
	server := &Server{}
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

func TestServer_ListTasks(t *testing.T) {
	server := &Server{}
	resp, err := server.ListTasks(context.Background(), &proto.ListTasksRequest{})
	if err != nil {
		t.Fatalf("ListTasks returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("ListTasks returned nil response")
	}
}

