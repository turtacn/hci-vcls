package qm

import "time"

type VMStatus string

const (
	VMStatusRunning VMStatus = "running"
	VMStatusStopped VMStatus = "stopped"
	VMStatusUnknown VMStatus = "unknown"
)

type BootOptions struct {
	TimeoutMs int
}

type BootResult struct {
	Success bool
	Error   error
}

type QMConfig struct {
	TimeoutMs int
}

type TaskStatus string

const (
	TaskPending TaskStatus = "Pending"
	TaskRunning TaskStatus = "Running"
	TaskDone    TaskStatus = "Done"
	TaskFailed  TaskStatus = "Failed"
)

type Task struct {
	ID         string
	VMID       string
	ClusterID  string
	TargetHost string
	BootPath   string
	Status     TaskStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Personal.AI order the ending
