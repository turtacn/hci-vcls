package ha

import (
	"context"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/qm"
)

type executorImpl struct {
	qmClient      qm.Client
	qmExecutor    qm.Executor
	mysqlAdapter  mysql.Adapter
	taskRepo      mysql.HATaskRepository
	metrics       metrics.Metrics
	log           logger.Logger
	batchInterval time.Duration
	failFast      bool
}

var _ Executor = &executorImpl{}

func NewExecutor(qmClient qm.Client, qmExecutor qm.Executor, mysqlAdapter mysql.Adapter, taskRepo mysql.HATaskRepository, m metrics.Metrics, log logger.Logger, batchInterval time.Duration, failFast bool) Executor {
	return &executorImpl{
		qmClient:      qmClient,
		qmExecutor:    qmExecutor,
		mysqlAdapter:  mysqlAdapter,
		taskRepo:      taskRepo,
		metrics:       m,
		log:           log,
		batchInterval: batchInterval,
		failFast:      failFast,
	}
}

func (e *executorImpl) Execute(ctx context.Context, plan *Plan) error {
	return e.ExecuteWithCallback(ctx, plan, nil)
}

func (e *executorImpl) ExecuteWithCallback(ctx context.Context, plan *Plan, onTaskDone func(VMTask)) error {
	if plan == nil || len(plan.Tasks) == 0 {
		return ErrInvalidPlan
	}

	start := time.Now()
	defer func() {
		if e.metrics != nil {
			e.metrics.ObserveHAExecutionDuration(plan.ClusterID, time.Since(start).Seconds())
		}
	}()

	// Group tasks by batch
	batches := make(map[int][]VMTask)
	for _, t := range plan.Tasks {
		if t.Status == TaskSkipped {
			continue
		}
		batches[t.BatchNo] = append(batches[t.BatchNo], t)
	}

	var hasFailure bool

	for i := 1; i <= plan.TotalBatches; i++ {
		batchTasks, ok := batches[i]
		if !ok || len(batchTasks) == 0 {
			continue
		}

		select {
		case <-ctx.Done():
			if e.log != nil {
				e.log.Warn("HA execution context canceled, assuming leadership lost")
			}
			return ErrLeadershipLost
		default:
		}

		batchHasFailure := false

		// In a real system, these would run in parallel goroutines.
		// For simplicity in the mock executor, we can run them sequentially or use a simple waitgroup.
		for _, task := range batchTasks {
			task.Status = TaskExecuting
			e.updateTaskStatus(ctx, task.ID, task.Status)

			// Execute via QM
			_, err := e.qmClient.StartVM(ctx, task.VMID, task.ClusterID, task.TargetHost, string(task.BootPath))

			if err != nil {
				task.Status = TaskFailed
				task.RetryCount++
				batchHasFailure = true
				if e.log != nil {
					e.log.Error("Failed to start VM via QM", "vm", task.VMID, "error", err)
				}
			} else {
				// Simulating WaitTask here for the executor logic
				// In a real system, StartVM might be async and we loop to check or call WaitTask.
				// For tests, StartVM returns immediately and we assume it's "Done" if no error.
				task.Status = TaskDone
			}

			e.updateTaskStatus(ctx, task.ID, task.Status)

			if e.metrics != nil {
				e.metrics.IncHATaskTotal(task.ClusterID, string(task.Status))
			}

			if onTaskDone != nil {
				onTaskDone(task)
			}
		}

		if batchHasFailure {
			hasFailure = true
			if e.failFast {
				return ErrPartialFailure
			}
		}

		if i < plan.TotalBatches {
			time.Sleep(e.batchInterval)
		}
	}

	if hasFailure {
		return ErrPartialFailure
	}

	return nil
}

func (e *executorImpl) updateTaskStatus(ctx context.Context, taskID string, status TaskStatus) {
	if e.taskRepo == nil {
		return
	}

	// Map internal ha.TaskStatus to mysql.TaskStatus
	var dbStatus mysql.TaskStatus
	switch status {
	case TaskPending:
		dbStatus = mysql.TaskPending
	case TaskExecuting:
		dbStatus = mysql.TaskRunning
	case TaskDone:
		dbStatus = mysql.TaskCompleted
	case TaskFailed:
		dbStatus = mysql.TaskFailed
	default:
		dbStatus = mysql.TaskFailed
	}

	err := e.taskRepo.UpdateStatus(ctx, taskID, dbStatus)
	if err != nil && e.log != nil {
		e.log.Warn("Failed to update task status in repo", "task", taskID, "error", err)
	}
}

// Personal.AI order the ending
