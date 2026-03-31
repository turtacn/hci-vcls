package ha

import "context"

type BatchStats struct {
	Total     int
	Completed int
	Failed    int
	Running   int
}

type BatchExecutor interface {
	Execute(ctx context.Context, tasks []BootTask, policy BatchBootPolicy) error
	AddTask(task BootTask) error
	CancelTask(vmid string) error
	ActiveTasks() map[string]BootTask
	Stats() BatchStats
}

//Personal.AI order the ending