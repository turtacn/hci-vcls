package grpc

import (
	"context"
	"testing"

	"github.com/turtacn/hci-vcls/pkg/api/proto"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/vcls"
)

type mockHAEngine struct {
	decision ha.HADecision
	err      error
	tasks    map[string]ha.BootTask
}

func (m *mockHAEngine) Start(ctx context.Context) error                         { return nil }
func (m *mockHAEngine) Stop() error                                             { return nil }
func (m *mockHAEngine) Evaluate(ctx context.Context, vmid string) (ha.HADecision, error) {
	return m.decision, m.err
}
func (m *mockHAEngine) Execute(ctx context.Context, decision ha.HADecision) error { return nil }
func (m *mockHAEngine) BatchBoot(ctx context.Context, decisions []ha.HADecision, policy ha.BatchBootPolicy) error {
	return nil
}
func (m *mockHAEngine) ClusterView() ha.ClusterView             { return ha.ClusterView{} }
func (m *mockHAEngine) ActiveTasks() map[string]ha.BootTask     { return m.tasks }
func (m *mockHAEngine) OnDecision(callback func(decision ha.HADecision)) {}

type mockFDMAgent struct {
	level fdm.DegradationLevel
	cv    fdm.ClusterView
}

func (m *mockFDMAgent) Start(ctx context.Context) error                         { return nil }
func (m *mockFDMAgent) Stop() error                                             { return nil }
func (m *mockFDMAgent) NodeStates() map[string]fdm.NodeState                    { return nil }
func (m *mockFDMAgent) LocalDegradationLevel() fdm.DegradationLevel             { return m.level }
func (m *mockFDMAgent) IsLeader() bool                                          { return false }
func (m *mockFDMAgent) LeaderNodeID() string                                    { return "" }
func (m *mockFDMAgent) OnNodeFailure(callback func(nodeID string))              {}
func (m *mockFDMAgent) OnDegradationChanged(callback func(level fdm.DegradationLevel)) {}
func (m *mockFDMAgent) ClusterView() fdm.ClusterView                            { return m.cv }

type mockVCLSAgent struct {
	caps []vcls.Capability
}

func (m *mockVCLSAgent) Start(ctx context.Context) error                         { return nil }
func (m *mockVCLSAgent) Stop() error                                             { return nil }
func (m *mockVCLSAgent) ClusterServiceState() vcls.ClusterServiceState           { return vcls.ServiceStateHealthy }
func (m *mockVCLSAgent) IsCapable(cap vcls.Capability) bool                      { return true }
func (m *mockVCLSAgent) RequireCapability(cap vcls.Capability) error             { return nil }
func (m *mockVCLSAgent) OnDegradationChanged(callback func(level fdm.DegradationLevel)) {}
func (m *mockVCLSAgent) ActiveCapabilities() []vcls.Capability                   { return m.caps }

func TestServer_Evaluate(t *testing.T) {
	haEngine := &mockHAEngine{
		decision: ha.HADecision{Action: ha.ActionBoot, TargetNode: "node-2", VMID: "100", Reason: "Ready"},
	}
	server := NewServer(haEngine, &mockFDMAgent{}, &mockVCLSAgent{})

	ctx := context.Background()
	req := &proto.EvaluateRequest{Vmid: "100"}

	resp, err := server.Evaluate(ctx, req)
	if err != nil {
		t.Fatalf("Expected evaluate to succeed, got %v", err)
	}

	if resp.Action != string(ha.ActionBoot) {
		t.Errorf("Expected ActionBoot, got %s", resp.Action)
	}
	if resp.TargetNode != "node-2" {
		t.Errorf("Expected node-2, got %s", resp.TargetNode)
	}
	if resp.Reason != "Ready" {
		t.Errorf("Expected Ready, got %s", resp.Reason)
	}
}

func TestServer_GetActiveTasks(t *testing.T) {
	haEngine := &mockHAEngine{
		tasks: map[string]ha.BootTask{
			"100": {VMID: "100", Status: ha.TaskRunning},
		},
	}
	server := NewServer(haEngine, &mockFDMAgent{}, &mockVCLSAgent{})

	ctx := context.Background()
	req := &proto.GetTasksRequest{}

	resp, err := server.GetActiveTasks(ctx, req)
	if err != nil {
		t.Fatalf("Expected GetActiveTasks to succeed, got %v", err)
	}

	if len(resp.Tasks) != 1 {
		t.Errorf("Expected 1 active task, got %d", len(resp.Tasks))
	}
	if resp.Tasks[0].Vmid != "100" || resp.Tasks[0].Status != string(ha.TaskRunning) {
		t.Errorf("Expected VM 100 to be running")
	}
}

func TestServer_GetClusterStatus(t *testing.T) {
	fdmAgent := &mockFDMAgent{
		cv: fdm.ClusterView{
			LeaderID: "node-1",
			Nodes: map[string]fdm.NodeState{
				"node-1": fdm.NodeStateAlive,
			},
		},
	}
	server := NewServer(&mockHAEngine{}, fdmAgent, &mockVCLSAgent{})

	ctx := context.Background()
	req := &proto.GetClusterStatusRequest{}

	resp, err := server.GetClusterStatus(ctx, req)
	if err != nil {
		t.Fatalf("Expected GetClusterStatus to succeed, got %v", err)
	}

	if resp.LeaderId != "node-1" {
		t.Errorf("Expected leader node-1, got %s", resp.LeaderId)
	}
	if state, ok := resp.NodeStates["node-1"]; !ok || state != string(fdm.NodeStateAlive) {
		t.Errorf("Expected node-1 to be alive")
	}
}

func TestServer_GetDegradation(t *testing.T) {
	fdmAgent := &mockFDMAgent{level: fdm.DegradationZK}
	server := NewServer(&mockHAEngine{}, fdmAgent, &mockVCLSAgent{})

	ctx := context.Background()
	req := &proto.GetDegradationRequest{}

	resp, err := server.GetDegradation(ctx, req)
	if err != nil {
		t.Fatalf("Expected GetDegradation to succeed, got %v", err)
	}

	if resp.Level != int32(fdm.DegradationZK) {
		t.Errorf("Expected level 1 (ZK), got %d", resp.Level)
	}
}

func TestServer_GetFullStatus(t *testing.T) {
	fdmAgent := &mockFDMAgent{level: fdm.DegradationNone}
	vclsAgent := &mockVCLSAgent{caps: []vcls.Capability{vcls.CapabilityHA, vcls.CapabilityDRS}}
	server := NewServer(&mockHAEngine{}, fdmAgent, vclsAgent)

	ctx := context.Background()
	req := &proto.GetFullStatusRequest{}

	resp, err := server.GetFullStatus(ctx, req)
	if err != nil {
		t.Fatalf("Expected GetFullStatus to succeed, got %v", err)
	}

	if resp.DegradationLevel != int32(fdm.DegradationNone) {
		t.Errorf("Expected level 0 (None), got %d", resp.DegradationLevel)
	}
	if len(resp.ActiveCapabilities) != 2 {
		t.Errorf("Expected 2 active capabilities, got %d", len(resp.ActiveCapabilities))
	}
	if resp.ActiveCapabilities[0] != string(vcls.CapabilityHA) {
		t.Errorf("Expected HA capability, got %s", resp.ActiveCapabilities[0])
	}
}

//Personal.AI order the ending