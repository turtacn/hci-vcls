package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/turtacn/hci-vcls/internal/app"
	"github.com/turtacn/hci-vcls/internal/config"
	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/heartbeat"
	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/api/rest"
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

	elector := election.NewMemoryElector(cfg.Node.NodeID)

	monitor := heartbeat.NewMemoryMonitor()
	evaluator := fdm.NewEvaluator()
	sm := statemachine.NewMachine()

	hbConfig := heartbeat.HeartbeatConfig{
		IntervalMs: int(cfg.Heartbeat.Interval.Milliseconds()),
		TimeoutMs:  int(cfg.Heartbeat.Timeout.Milliseconds()),
		NodeID:     cfg.Node.NodeID,
	}

	appLogger := logger.NewLogger(cfg.Log.Level, cfg.Log.Format)

	hbService := heartbeat.NewService(hbConfig, monitor, elector, evaluator, sm, m, appLogger)

	store := vcls.NewMemoryStore()
	vclsService := vcls.NewService(store, cfsClient, vmRepo, witClient, nil, nil, m, appLogger)

	planner := ha.NewPlanner()
	executor := ha.NewExecutor(qmClient, taskRepo, m, appLogger, cfg.HA.BatchInterval, cfg.HA.FailFast)

	fdmConfig := fdm.FDMConfig{
		NodeID:          cfg.Node.NodeID,
		ClusterID:       cfg.Node.ClusterID,
		ProbeIntervalMs: int(cfg.FDM.EvalInterval.Milliseconds()),
	}
	fdmAgent := fdm.NewAgent(fdmConfig, &dummyProber{}, elector, appLogger, m)

	appSvc := app.NewService(cfg, log, m, elector, hbService, vclsService, planner, executor, sm, vmRepo, planRepo, fdmAgent)

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

// Personal.AI order the ending
