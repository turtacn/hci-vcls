package helpers

import (
	"fmt"
	"net"
	"path/filepath"
	"runtime"
	"time"

	"go.uber.org/zap"

	"github.com/turtacn/hci-vcls/internal/app"
	"github.com/turtacn/hci-vcls/internal/config"
	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/heartbeat"
	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/qm"
	"github.com/turtacn/hci-vcls/pkg/statemachine"
	"github.com/turtacn/hci-vcls/pkg/vcls"
	"github.com/turtacn/hci-vcls/pkg/witness"
)

func ProjectRoot() string {
	_, b, _, _ := runtime.Caller(0)
	// Return up to the root folder. E.g. b = /.../test/e2e/helpers/helpers.go -> ...
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(b))))
}

func FreePort() int {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		panic(err)
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port
}

func DefaultTestConfig() *config.Config {
	cfg := &config.Config{}
	cfg.Node.NodeID = "test-node-1"
	cfg.Node.ClusterID = "test-cluster"
	cfg.Node.HostIP = "127.0.0.1"

	cfg.Server.HTTPAddr = fmt.Sprintf(":%d", FreePort())
	cfg.Server.GRPCAddr = fmt.Sprintf(":%d", FreePort())

	cfg.Heartbeat.Interval = 10 * time.Millisecond
	cfg.Heartbeat.Timeout = 50 * time.Millisecond

	cfg.Election.LeaseTTL = 10 * time.Second

	cfg.HA.BatchSize = 5
	cfg.HA.FailFast = false

	return cfg
}

type TestApp struct {
	Service   *app.Service
	CFS       *cfs.MemoryClient
	QM        *qm.MemoryClient
	Witness   *witness.MemoryClient
	Repo      *mysql.MemoryVMRepository
	TaskRepo  *mysql.MemoryHATaskRepository
	PlanRepo  *mysql.MemoryPlanRepository
	Monitor   heartbeat.Monitor
	Elector   election.Elector
	HBService *heartbeat.HeartbeatService
}

func NewTestService(cfg *config.Config) *TestApp {
	log := zap.NewNop()
	m := metrics.NewNoopMetrics()

	cfsClient := cfs.NewMemoryClient()
	qmClient := qm.NewMemoryClient(0.0, 1*time.Millisecond)
	witClient := witness.NewMemoryClient()

	vmRepo := mysql.NewMemoryVMRepository()
	taskRepo := mysql.NewMemoryHATaskRepository()
	planRepo := mysql.NewMemoryPlanRepository()

	elector := election.NewMemoryElector(cfg.Node.NodeID)
	evaluator := fdm.NewEvaluator()
	sm := statemachine.NewMachine()

	monitor := heartbeat.NewMemoryMonitor()

	hbConfig := heartbeat.HeartbeatConfig{
		IntervalMs: int(cfg.Heartbeat.Interval.Milliseconds()),
		TimeoutMs:  int(cfg.Heartbeat.Timeout.Milliseconds()),
		NodeID:     cfg.Node.NodeID,
	}

	mockHB := heartbeat.NewHeartbeater(hbConfig, nil)
	hbService := heartbeat.NewService(hbConfig, mockHB, monitor, elector, evaluator, sm, m, nil)

	store := vcls.NewMemoryStore()
	vclsService := vcls.NewService(store, cfsClient, vmRepo, witClient, nil, nil, m, nil)

	planner := ha.NewPlanner()
	executor := ha.NewExecutor(qmClient, nil, nil, taskRepo, m, nil, 0, cfg.HA.FailFast)

	appSvc := app.NewService(cfg, log, m, elector, hbService, vclsService, planner, executor, sm, vmRepo, planRepo, nil)

	return &TestApp{
		Service:   appSvc,
		CFS:       cfsClient,
		QM:        qmClient,
		Witness:   witClient,
		Repo:      vmRepo,
		TaskRepo:  taskRepo,
		PlanRepo:  planRepo,
		Monitor:   monitor,
		Elector:   elector,
		HBService: hbService,
	}
}

