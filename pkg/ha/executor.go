package ha

import (
	"context"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
	"github.com/turtacn/hci-vcls/pkg/metrics"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/qm"
)

type cacheProvider interface {
	GetComputeMeta(ctx context.Context, vmid string) (interface{}, error)
}

// stateProvider allows executor to query holistic cluster state without cycling imports
type stateProvider interface {
	CurrentLevel() string
}

type executorImpl struct {
	qmClient      qm.Client
	qmExecutor    qm.Executor
	mysqlAdapter  mysql.Adapter
	taskRepo      mysql.HATaskRepository
	metrics       metrics.Metrics
	log           logger.Logger
	batchInterval time.Duration
	failFast      bool
	cache         cacheProvider
	stateMachine  stateProvider
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

func (e *executorImpl) SetCache(c cacheProvider) {
	e.cache = c
}

func (e *executorImpl) SetStateMachine(sm stateProvider) {
	e.stateMachine = sm
}

// ExecuteWithPlan handles untyped plan interface to avoid dependency loops from fdm agent mock calls
func (e *executorImpl) ExecuteWithPlan(ctx context.Context, planInterface interface{}) error {
	plan, ok := planInterface.(*Plan)
	if !ok {
		return ErrInvalidPlan
	}
	return e.ExecuteWithCallback(ctx, plan, nil)
}

func (e *executorImpl) Execute(ctx context.Context, plan *Plan) error {
	return e.ExecuteWithCallback(ctx, plan, nil)
}

func (e *executorImpl) ExecuteWithCallback(ctx context.Context, plan *Plan, onTaskDone func(VMTask)) error {
	if plan == nil || len(plan.Tasks) == 0 {
		return ErrInvalidPlan
	}

	currentLevel := e.getCurrentLevel(plan)

	if err := e.validateExecutionGate(plan, currentLevel); err != nil {
		return err
	}

	start := time.Now()
	defer func() {
		if e.metrics != nil {
			e.metrics.ObserveHAExecutionDuration(plan.ClusterID, time.Since(start).Seconds())
		}
	}()

	batches := e.groupTasksByBatch(plan)
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

		batchHasFailure := e.executeBatch(ctx, batchTasks, plan, currentLevel, onTaskDone)
		if batchHasFailure {
			hasFailure = true
			if e.failFast {
				return ErrPartialFailure
			}
		}

		if i < plan.TotalBatches {
			select {
			case <-ctx.Done():
				if e.log != nil {
					e.log.Warn("HA execution context canceled during batch interval, assuming leadership lost")
				}
				return ErrLeadershipLost
			case <-time.After(e.batchInterval):
			}
		}
	}

	if hasFailure {
		return ErrPartialFailure
	}

	return nil
}

func (e *executorImpl) getCurrentLevel(plan *Plan) string {
	currentLevel := "None"
	if e.stateMachine != nil {
		currentLevel = e.stateMachine.CurrentLevel()
	} else if plan.Degradation != "" {
		currentLevel = plan.Degradation
	}
	return currentLevel
}

func (e *executorImpl) validateExecutionGate(plan *Plan, level string) error {
	// Refactored to not strictly use magic strings if possible, but currently we receive level as string.
	// Assume "Critical" maps to Isolated.
	if level == "Critical" {
		if e.log != nil {
			e.log.Warn("HA execution skipped: cluster degradation level is Critical (Isolated)", "cluster", plan.ClusterID)
		}
		return ErrSkippedIsolated
	}
	return nil
}

func (e *executorImpl) groupTasksByBatch(plan *Plan) map[int][]VMTask {
	batches := make(map[int][]VMTask)
	for _, t := range plan.Tasks {
		if t.Status == TaskSkipped {
			continue
		}
		batches[t.BatchNo] = append(batches[t.BatchNo], t)
	}
	return batches
}

func (e *executorImpl) executeBatch(ctx context.Context, batchTasks []VMTask, plan *Plan, currentLevel string, onTaskDone func(VMTask)) bool {
	batchHasFailure := false

	// Execute sequentially within the batch for simplicity (or can be parallelized later)
	for _, task := range batchTasks {
		task.Status = TaskExecuting
		e.updateTaskStatus(ctx, task.ID, task.Status)

		// 1. Claim Boot
		tx, claimErr := e.claimBoot(ctx, &task, plan, currentLevel)
		if claimErr != nil {
			batchHasFailure = true
			continue
		}

		// 2. Load Cached Meta if necessary
		meta, metaErr := e.loadCachedMeta(ctx, &task)
		if metaErr != nil && e.log != nil {
			e.log.Warn("Failed to read from cache during fallback", "vm", task.VMID, "error", metaErr)
			// Decide if cache miss is terminal or not. For now, log warning and proceed (might fail qm start if meta is critical).
			// If meta is strictly required for this boot path, we could fail early here.
		}

		// 3. Start VM via QM
		err := e.startVM(ctx, &task, meta)

		// 4. Finalize Task (Handle idempotency and errors)
		terminalErr := e.finalizeTask(ctx, tx, &task, err)
		if terminalErr != nil {
			batchHasFailure = true
		}

		if onTaskDone != nil {
			onTaskDone(task)
		}
	}

	return batchHasFailure
}

func (e *executorImpl) claimBoot(ctx context.Context, task *VMTask, plan *Plan, currentLevel string) (mysql.Tx, error) {
	if e.mysqlAdapter == nil {
		return nil, nil // No adapter, proceed without claim (or mock)
	}

	// Avoid split brain if MySQL is definitely unavailable
	if currentLevel == "Major" && e.mysqlAdapter.Health().State == mysql.MySQLStateUnavailable {
		task.Status = TaskFailed
		task.Reason = "MySQL unavailable, cannot acquire optimistic lock"
		e.updateTaskStatus(ctx, task.ID, task.Status)
		if e.log != nil {
			e.log.Error("MySQL unavailable, skipping boot to avoid split-brain", "vm", task.VMID)
		}
		return nil, ErrPartialFailure
	}

	tx, err := e.mysqlAdapter.BeginTx()
	if err != nil {
		task.Status = TaskFailed
		task.Reason = "failed to start transaction for optimistic lock"
		e.updateTaskStatus(ctx, task.ID, task.Status)
		if e.log != nil {
			e.log.Warn("Failed to start tx for boot claim", "vm", task.VMID, "error", err)
		}
		return nil, err
	}

	claimErr := tx.ClaimBoot(mysql.BootClaim{
		VMID:       task.VMID,
		Token:      plan.ID,
		TargetNode: task.TargetHost,
	})

	if claimErr != nil {
		_ = tx.Rollback()
		task.Status = TaskFailed
		task.Reason = "optimistic lock failed"
		e.updateTaskStatus(ctx, task.ID, task.Status)
		if e.log != nil {
			e.log.Warn("Failed to claim boot", "vm", task.VMID, "error", claimErr)
		}
		return nil, claimErr
	}

	return tx, nil
}

func (e *executorImpl) loadCachedMeta(ctx context.Context, task *VMTask) (interface{}, error) {
	// If path relies on minority or cache
	if task.BootPath == BootPathMinority && e.cache != nil {
		meta, err := e.cache.GetComputeMeta(ctx, task.VMID)
		// In a complete implementation, store meta into some struct expected by qm
		return meta, err
	}
	return nil, nil
}

func (e *executorImpl) startVM(ctx context.Context, task *VMTask, meta interface{}) error {
	// Assuming StartVM doesn't accept meta yet, we pass what we can
	_, err := e.qmClient.StartVM(ctx, task.VMID, task.ClusterID, task.TargetHost, string(task.BootPath))
	return err
}

func (e *executorImpl) finalizeTask(ctx context.Context, tx mysql.Tx, task *VMTask, startErr error) error {
	if startErr != nil {
		// Idempotency: "already running" is considered success
		if startErr == qm.ErrVMAlreadyRunning {
			task.Status = TaskDone
			task.Reason = "already running"
			if e.log != nil {
				e.log.Info("VM already running, treating as success", "vm", task.VMID)
			}
			if tx != nil {
				_ = tx.Commit() // we successfully claimed and it's running
			}
		} else {
			// Failure
			task.Status = TaskFailed
			task.Reason = startErr.Error()
			task.RetryCount++
			if e.log != nil {
				e.log.Error("Failed to start VM via QM", "vm", task.VMID, "error", startErr)
			}
			if tx != nil {
				// We claimed but failed to start, so rollback the claim
				// TODO: In adapter level, tx.Rollback() should release the claim or mark it as failed if DB state was updated explicitly.
				_ = tx.Rollback()
			}
		}
	} else {
		task.Status = TaskDone
		task.Reason = "success"
		if tx != nil {
			_ = tx.Commit()
		}
	}

	e.updateTaskStatus(ctx, task.ID, task.Status)

	if e.metrics != nil {
		e.metrics.IncHATaskTotal(task.ClusterID, string(task.Status))
	}

	if task.Status == TaskFailed {
		return startErr
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
