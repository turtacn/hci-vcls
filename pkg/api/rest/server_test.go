package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/turtacn/hci-vcls/internal/app"
	"github.com/turtacn/hci-vcls/internal/config"
	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/statemachine"
)

type mockAgent struct{}

func (m *mockAgent) Start(ctx context.Context) error                 { return nil }
func (m *mockAgent) Stop() error                                     { return nil }
func (m *mockAgent) LocalDegradationLevel() fdm.DegradationLevel     { return fdm.DegradationNone }
func (m *mockAgent) OnDegradationChanged(func(fdm.DegradationLevel)) {}
func (m *mockAgent) OnNodeFailure(func(string))                      {}
func (m *mockAgent) NodeStates() map[string]fdm.NodeState            { return nil }
func (m *mockAgent) IsLeader() bool                                  { return true }
func (m *mockAgent) LeaderNodeID() string                            { return "" }
func (m *mockAgent) ClusterView() fdm.ClusterView                    { return fdm.ClusterView{} }

func setupTestRouter() (*gin.Engine, *app.Service) {
	gin.SetMode(gin.TestMode)

	log := zap.NewNop()
	cfg := &config.Config{}
	cfg.Node.ClusterID = "test-cluster"

	elector := election.NewMemoryElector("test-node", nil)
	elector.SetNodesCount(1) // Single node cluster becomes leader immediately
	_ = elector.Campaign(context.Background())
	m := metrics.NewNoopMetrics()

	sm := statemachine.NewMachine(nil)
	vmRepo := mysql.NewMemoryVMRepository()
	planRepo := mysql.NewMemoryPlanRepository()

	fdmAgent := &mockAgent{}

	svc := app.NewService(cfg, log, m, elector, nil, nil, nil, nil, sm, vmRepo, planRepo, fdmAgent, nil, nil)

	handler := NewHandler(svc, log)
	router := gin.Default()
	handler.RegisterRoutes(router)

	return router, svc
}

func TestVersion(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/version", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["version"] != "1.0.0" {
		t.Errorf("Expected version 1.0.0, got %s", resp["version"])
	}
}

func TestStatus(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/status", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["IsLeader"] != true {
		t.Errorf("Expected IsLeader to be true")
	}
}

func TestDegradation(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/degradation", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["level"] != string(fdm.DegradationNone) {
		t.Errorf("Expected degradation level %s, got %s", fdm.DegradationNone, resp["level"])
	}
}

func TestEvaluateHA(t *testing.T) {
	router, _ := setupTestRouter()

	body := []byte(`{"cluster_id": "test-cluster"}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ha/evaluate", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	// Since threshold is None and we need Major/Critical, it will return BELOW_THRESHOLD
	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["code"] != "BELOW_THRESHOLD" {
		t.Errorf("Expected BELOW_THRESHOLD, got %v", resp["code"])
	}
}

func TestEvaluateHA_MissingCluster(t *testing.T) {
	router, _ := setupTestRouter()

	body := []byte(`{}`)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/ha/evaluate", bytes.NewBuffer(body))
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400, got %d", w.Code)
	}
}

func TestListTasks(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/ha/tasks", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	tasks, ok := resp["tasks"].([]interface{})
	if !ok || len(tasks) != 0 {
		t.Errorf("Expected empty tasks list")
	}
}

func TestServerStartStop(t *testing.T) {
	h := NewHandler(nil, zap.NewNop())
	srv := NewServer(":0", h) // random port

	go func() {
		_ = srv.Start()
	}()

	time.Sleep(50 * time.Millisecond)

	err := srv.Stop(context.Background())
	if err != nil {
		t.Fatalf("Failed to stop server: %v", err)
	}
}

