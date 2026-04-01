package qm

import (
	"context"
	"time"
)

type Client interface {
	StartVM(ctx context.Context, vmID, clusterID, targetHost, bootPath string) (*Task, error)
	StopVM(ctx context.Context, vmID, clusterID string) (*Task, error)
	GetTask(ctx context.Context, taskID string) (*Task, error)
	WaitTask(ctx context.Context, taskID string, timeout time.Duration) (*Task, error)
}

// Personal.AI order the ending
