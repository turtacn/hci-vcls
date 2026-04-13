package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/turtacn/hci-vcls/internal/app"
	"github.com/turtacn/hci-vcls/internal/app/audit"
	"github.com/turtacn/hci-vcls/internal/config"
	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/heartbeat"
	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/api/rest"
	"github.com/turtacn/hci-vcls/pkg/cache"
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

type dummyProber struct{}

func (d *dummyProber) ProbeL0(ctx context.Context) fdm.ProbeResult {
	return fdm.ProbeResult{Success: true}
}
func (d *dummyProber) ProbeL1(ctx context.Context) fdm.ProbeResult {
	return fdm.ProbeResult{Success: true}
}
func (d *dummyProber) ProbeL2(ctx context.Context) fdm.ProbeResult {
	return fdm.ProbeResult{Success: true}
}
func (d *dummyProber) ProbeAll(ctx context.Context) map[fdm.HeartbeatLevel]fdm.ProbeResult {
	return map[fdm.HeartbeatLevel]fdm.ProbeResult{
		fdm.HeartbeatL0: {Success: true},
	}
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HCI vCLS Server",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServe(appConfig)
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func runServe(cfg *config.Config) error {
	var log *zap.Logger
	if cfg.Log.Level == "debug" {
		log, _ = zap.NewDevelopment()
	} else {
		log, _ = zap.NewProduction()
	}
	defer log.Sync()

	m := metrics.NewNoopMetrics()

	cfsClient := cfs.NewMemoryClient()
	qmClient := qm.NewMemoryClient(0.0, 0)
	witClient := witness.NewMemoryClient()

	vmRepo := mysql.NewMemoryVMRepository()
	taskRepo := mysql.NewMemoryHATaskRepository()
	planRepo := mysql.NewMemoryPlanRepository()

	termStore, err := election.NewGobTermStore("/var/lib/hci-vcls/election/term.gob")
	if err != nil {
		return fmt.Errorf("failed to initialize term store: %v", err)
	}
	elector := election.NewMemoryElector(cfg.Node.NodeID, termStore)

	monitor := heartbeat.NewMemoryMonitor()
	evaluator := fdm.NewEvaluator()
	sm := statemachine.NewMachine(m)

	hbConfig := heartbeat.HeartbeatConfig{
		IntervalMs: int(cfg.Heartbeat.Interval.Milliseconds()),
		TimeoutMs:  int(cfg.Heartbeat.Timeout.Milliseconds()),
		NodeID:     cfg.Node.NodeID,
	}

	appLogger := logger.NewLogger(cfg.Log.Level, cfg.Log.Format)

	udpHeartbeater := heartbeat.NewUDPHeartbeater(hbConfig)
	hbService := heartbeat.NewService(hbConfig, udpHeartbeater, monitor, elector, evaluator, sm, m, appLogger)

	store := vcls.NewMemoryStore()
	// NOTE(phase06): wire actual CacheManager instance here when the
	// cmd-layer cache bootstrap is refactored. Passing nil preserves
	// phase04 behavior (no background cache tracking).
	vclsService := vcls.NewServiceWithCacheManager(store, cfsClient, vmRepo, witClient, nil, nil, nil, m, appLogger)

	// Init true minority boot adapters if MySQL DSN is provided, else use placeholders
	qmExecutor := qm.NewQMAdapter("/usr/sbin/qm")

	var mysqlAdapter mysql.Adapter
	if cfg.MySQL.DSN != "" {
		mConfig := mysql.MySQLConfig{
			DSN:          cfg.MySQL.DSN,
			MaxOpenConns: 10,
			MaxIdleConns: 5,
		}
		mysqlAdapter, _ = mysql.NewAdapter(mConfig, appLogger)
	}

	planner := ha.NewPlanner()
	executor := ha.NewExecutor(qmClient, qmExecutor, mysqlAdapter, taskRepo, m, appLogger, cfg.HA.BatchInterval, cfg.HA.FailFast)

	// Wire cache fallback path (phase05 T5) — enables minority boot path to
	// read VM metadata from local snapshot when CFS is read-only.
	// NOTE(phase06): wire actual CacheManager instance here when the
	// cmd-layer cache bootstrap is refactored. Passing nil cacheManager ensures safe placeholder.
	var cacheManager cache.CacheManager = nil
	if cacheManager != nil {
		if ce, ok := executor.(interface{ SetCache(c ha.CacheProvider) }); ok {
			ce.SetCache(app.NewHACacheAdapter(cacheManager))
		}
	}

	if sm != nil {
		if se, ok := executor.(interface{ SetStateMachine(sm ha.StateProvider) }); ok {
			se.SetStateMachine(app.NewStateMachineAdapter(sm))
		}
	}

	fdmConfig := fdm.FDMConfig{
		NodeID:          cfg.Node.NodeID,
		ClusterID:       cfg.Node.ClusterID,
		ProbeIntervalMs: int(cfg.FDM.EvalInterval.Milliseconds()),
	}
	fdmAgent := fdm.NewAgent(fdmConfig, &dummyProber{}, elector, appLogger, m)

	sweeperConfig := ha.SweeperConfig{
		StaleThreshold: 10 * time.Minute,
		ScanInterval:   2 * time.Minute,
	}
	isLeaderFunc := func() bool {
		return fdmAgent.IsLeader()
	}
	sweeper := ha.NewSweeper(sweeperConfig, mysqlAdapter, isLeaderFunc, m, appLogger)

	// Start the sweeper in background
	if err := sweeper.Start(context.Background()); err != nil {
		log.Error("Failed to start sweeper", zap.Error(err))
	}

	planCache, _ := app.NewFSPlanCache("/var/lib/hci-vcls/plans")

	auditLogger, err := audit.NewJSONLAuditLogger("/var/lib/hci-vcls/audit")
	if err == nil {
		if ea, ok := executor.(interface{ SetAudit(a ha.AuditSink) }); ok {
			ea.SetAudit(app.NewHAAuditAdapter(auditLogger))
		}
	}

	appSvc := app.NewService(cfg, log, m, elector, hbService, vclsService, planner, executor, sm, vmRepo, planRepo, fdmAgent, sweeper, planCache)

	handler := rest.NewHandler(appSvc, log)
	restServer := rest.NewServer(cfg.Server.HTTPAddr, handler)

	if err := hbService.Start(); err != nil {
		return fmt.Errorf("failed to start heartbeat service: %v", err)
	}

	go func() {
		if err := restServer.Start(); err != nil {
			log.Error("REST server error", zap.Error(err))
		}
	}()
	log.Info("Server started", zap.String("http_addr", cfg.Server.HTTPAddr))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Shutting down server...")

	hbService.Stop()
	return nil
}
