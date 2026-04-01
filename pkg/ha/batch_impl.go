package ha

import (
	"context"

	"github.com/turtacn/hci-vcls/pkg/qm"
)

type batchExecutorImpl struct {
	qmExecutor qm.Executor
	tasks      map[string]BootTask
	stats      BatchStats
}

func NewBatchExecutor(qmExecutor qm.Executor) BatchExecutor {
	return &batchExecutorImpl{
		qmExecutor: qmExecutor,
		tasks:      make(map[string]BootTask),
		stats:      BatchStats{},
	}
}

func (b *batchExecutorImpl) Execute(ctx context.Context, tasks []BootTask, policy BatchBootPolicy) error {
	for _, t := range tasks {
		_ = b.AddTask(t)
	}

	for vmid, t := range b.tasks {
		if t.Status == TaskPending {
			opts := qm.BootOptions{TimeoutMs: 1000} // Example config
			result := b.qmExecutor.Start(ctx, vmid, opts)

			if result.Success {
				t.Status = TaskRunning
				b.stats.Running++
			} else {
				t.Status = TaskFailed
				t.LastError = result.Error
				b.stats.Failed++
			}
			b.tasks[vmid] = t
		}
	}
	return nil
}

func (b *batchExecutorImpl) AddTask(task BootTask) error {
	b.tasks[task.VMID] = task
	b.stats.Total++
	return nil
}

func (b *batchExecutorImpl) CancelTask(vmid string) error {
	if t, ok := b.tasks[vmid]; ok {
		t.Status = TaskCompleted // Simplified cancellation
		b.tasks[vmid] = t
		b.stats.Completed++
	}
	return nil
}

func (b *batchExecutorImpl) ActiveTasks() map[string]BootTask {
	active := make(map[string]BootTask)
	for id, t := range b.tasks {
		if t.Status == TaskPending || t.Status == TaskRunning {
			active[id] = t
		}
	}
	return active
}

func (b *batchExecutorImpl) Stats() BatchStats {
	return b.stats
}

//Personal.AI order the ending