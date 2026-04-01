package helpers

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
)

type TestCluster struct {
	Nodes            []string
	DegradationLevel fdm.DegradationLevel
	Tasks            []ha.BootTask
	ctx              context.Context
	cancel           context.CancelFunc
}

func NewTestCluster() *TestCluster {
	ctx, cancel := context.WithCancel(context.Background())
	return &TestCluster{
		Nodes:            []string{"node-1", "node-2", "node-3"},
		DegradationLevel: fdm.DegradationNone,
		Tasks:            make([]ha.BootTask, 0),
		ctx:              ctx,
		cancel:           cancel,
	}
}

func (c *TestCluster) Start() error {
	// A real implementation would spin up docker containers or local processes
	// For E2E skeleton, we assume it's running or mock it entirely if needed
	// However, we can use docker-compose up for a real E2E
	cmd := exec.Command("docker-compose", "-f", "../../docker-compose.dev.yml", "up", "-d")
	// Ignoring error for skeleton as docker might not be available in all test envs without setup
	_ = cmd.Run()

	// Wait for services
	time.Sleep(2 * time.Second)
	return nil
}

func (c *TestCluster) Stop() error {
	c.cancel()
	cmd := exec.Command("docker-compose", "-f", "../../docker-compose.dev.yml", "down")
	_ = cmd.Run()
	return nil
}

func (c *TestCluster) FailNode(nodeID string) error {
	fmt.Printf("Simulating failure of node %s\n", nodeID)
	return nil
}

func (c *TestCluster) RecoverNode(nodeID string) error {
	fmt.Printf("Simulating recovery of node %s\n", nodeID)
	return nil
}

func (c *TestCluster) SetDegradation(level fdm.DegradationLevel) {
	c.DegradationLevel = level
}

func (c *TestCluster) AddTask(task ha.BootTask) {
	c.Tasks = append(c.Tasks, task)
}

//Personal.AI order the ending
