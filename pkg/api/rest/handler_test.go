package rest

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/vcls"
)

type mockHAEngine struct {
	decision ha.HADecision
	err      error
	tasks    map[string]ha.BootTask
}

func (m *mockHAEngine) Start(ctx context.Context) error { return nil }
func (m *mockHAEngine) Stop() error                     { return nil }
func (m *mockHAEngine) Evaluate(ctx context.Context, vmid string) (ha.HADecision, error) {
	return m.decision, m.err
}
func (m *mockHAEngine) Execute(ctx context.Context, decision ha.HADecision) error { return nil }
func (m *mockHAEngine) BatchBoot(ctx context.Context, decisions []ha.HADecision, policy ha.BatchBootPolicy) error {
	return nil
}
func (m *mockHAEngine) ClusterView() ha.ClusterView                      { return ha.ClusterView{} }
func (m *mockHAEngine) ActiveTasks() map[string]ha.BootTask              { return m.tasks }
func (m *mockHAEngine) OnDecision(callback func(decision ha.HADecision)) {}

type mockFDMAgent struct {
	level fdm.DegradationLevel
	cv    fdm.ClusterView
}

func (m *mockFDMAgent) Start(ctx context.Context) error                                { return nil }
func (m *mockFDMAgent) Stop() error                                                    { return nil }
func (m *mockFDMAgent) NodeStates() map[string]fdm.NodeState                           { return nil }
func (m *mockFDMAgent) LocalDegradationLevel() fdm.DegradationLevel                    { return m.level }
func (m *mockFDMAgent) IsLeader() bool                                                 { return false }
func (m *mockFDMAgent) LeaderNodeID() string                                           { return "" }
func (m *mockFDMAgent) OnNodeFailure(callback func(nodeID string))                     {}
func (m *mockFDMAgent) OnDegradationChanged(callback func(level fdm.DegradationLevel)) {}
func (m *mockFDMAgent) ClusterView() fdm.ClusterView                                   { return m.cv }

type mockVCLSAgent struct{}

func (m *mockVCLSAgent) Start(ctx context.Context) error { return nil }
func (m *mockVCLSAgent) Stop() error                     { return nil }
func (m *mockVCLSAgent) ClusterServiceState() vcls.ClusterServiceState {
	return vcls.ServiceStateHealthy
}
func (m *mockVCLSAgent) IsCapable(cap vcls.Capability) bool                             { return true }
func (m *mockVCLSAgent) RequireCapability(cap vcls.Capability) error                    { return nil }
func (m *mockVCLSAgent) OnDegradationChanged(callback func(level fdm.DegradationLevel)) {}
func (m *mockVCLSAgent) ActiveCapabilities() []vcls.Capability                          { return nil }

func setupRouter(haEngine ha.HAEngine, fdmAgent fdm.Agent, vclsAgent vcls.Agent) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	handler := NewHandler(haEngine, fdmAgent, vclsAgent)
	handler.RegisterRoutes(router)
	return router
}

func TestGetStatus(t *testing.T) {
	fdmAgent := &mockFDMAgent{
		cv: fdm.ClusterView{
			LeaderID: "node-1",
			Nodes: map[string]fdm.NodeState{
				"node-1": fdm.NodeStateAlive,
			},
		},
	}
	router := setupRouter(&mockHAEngine{}, fdmAgent, &mockVCLSAgent{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/status", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["leader_id"] != "node-1" {
		t.Errorf("Expected leader node-1, got %s", response["leader_id"])
	}
}

func TestGetDegradation(t *testing.T) {
	fdmAgent := &mockFDMAgent{level: fdm.DegradationZK}
	router := setupRouter(&mockHAEngine{}, fdmAgent, &mockVCLSAgent{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/degradation", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["level"].(float64) != float64(fdm.DegradationZK) {
		t.Errorf("Expected level 1 (ZK), got %f", response["level"])
	}
}

func TestGetTasks(t *testing.T) {
	haEngine := &mockHAEngine{
		tasks: map[string]ha.BootTask{
			"100": {VMID: "100", Status: ha.TaskRunning},
		},
	}
	router := setupRouter(haEngine, &mockFDMAgent{}, &mockVCLSAgent{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/tasks", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	tasks := response["tasks"].(map[string]interface{})
	if task, ok := tasks["100"].(map[string]interface{}); !ok || task["Status"] != string(ha.TaskRunning) {
		t.Errorf("Expected VM 100 to be running")
	}
}

func TestEvaluate(t *testing.T) {
	haEngine := &mockHAEngine{
		decision: ha.HADecision{Action: ha.ActionBoot, TargetNode: "node-2", VMID: "100", Reason: "Ready"},
	}
	router := setupRouter(haEngine, &mockFDMAgent{}, &mockVCLSAgent{})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/evaluate/100", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var decision ha.HADecision
	err := json.Unmarshal(w.Body.Bytes(), &decision)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if decision.Action != ha.ActionBoot {
		t.Errorf("Expected ActionBoot, got %v", decision.Action)
	}
	if decision.TargetNode != "node-2" {
		t.Errorf("Expected node-2, got %s", decision.TargetNode)
	}
}

//Personal.AI order the ending
