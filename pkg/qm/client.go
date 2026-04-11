package qm

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var (
	ErrNotFound = errors.New("task not found")
	ErrTimeout  = errors.New("wait task timeout")
)

type MemoryClient struct {
	mu           sync.RWMutex
	tasks        map[string]*Task
	taskCount    int
	failureRatio float64 // 0.0 to 1.0
	delay        time.Duration
}

var _ Client = &MemoryClient{}

func NewMemoryClient(failureRatio float64, delay time.Duration) *MemoryClient {
	return &MemoryClient{
		tasks:        make(map[string]*Task),
		failureRatio: failureRatio,
		delay:        delay,
	}
}

func (c *MemoryClient) StartVM(ctx context.Context, vmID, clusterID, targetHost, bootPath string) (*Task, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.taskCount++
	taskID := fmt.Sprintf("task-%d", c.taskCount)

	task := &Task{
		ID:         taskID,
		VMID:       vmID,
		ClusterID:  clusterID,
		TargetHost: targetHost,
		BootPath:   bootPath,
		Status:     TaskPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	c.tasks[taskID] = task

	// Copy taskID to avoid capturing loop/reference variables improperly
	go c.simulateTaskExecution(taskID)

	// Return a copy to avoid races
	taskCopy := *task
	return &taskCopy, nil
}

func (c *MemoryClient) StopVM(ctx context.Context, vmID, clusterID string) (*Task, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.taskCount++
	taskID := fmt.Sprintf("task-stop-%d", c.taskCount)

	task := &Task{
		ID:        taskID,
		VMID:      vmID,
		ClusterID: clusterID,
		Status:    TaskPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	c.tasks[taskID] = task

	go c.simulateTaskExecution(taskID)

	taskCopy := *task
	return &taskCopy, nil
}

func (c *MemoryClient) GetTask(ctx context.Context, taskID string) (*Task, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	task, exists := c.tasks[taskID]
	if !exists {
		return nil, ErrNotFound
	}

	// Return a copy to prevent race
	tCopy := *task
	return &tCopy, nil
}

func (c *MemoryClient) WaitTask(ctx context.Context, taskID string, timeout time.Duration) (*Task, error) {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return nil, ErrTimeout
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(10 * time.Millisecond):
			task, err := c.GetTask(ctx, taskID)
			if err != nil {
				return nil, err
			}
			if task.Status == TaskDone || task.Status == TaskFailed {
				return task, nil
			}
		}
	}
}

func (c *MemoryClient) simulateTaskExecution(taskID string) {
	time.Sleep(c.delay)

	c.mu.Lock()
	defer c.mu.Unlock()

	task, exists := c.tasks[taskID]
	if !exists {
		return
	}

	// Move to running
	task.Status = TaskRunning
	task.UpdatedAt = time.Now()

	time.Sleep(c.delay)

	// Complete or fail
	if rand.Float64() < c.failureRatio {
		task.Status = TaskFailed
	} else {
		task.Status = TaskDone
	}
	task.UpdatedAt = time.Now()
}

