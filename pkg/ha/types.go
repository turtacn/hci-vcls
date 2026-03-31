package ha

import "github.com/turtacn/hci-vcls/pkg/fdm"

type BootPath string

const (
	BootPathZK     BootPath = "ZK"
	BootPathCFS    BootPath = "CFS"
	BootPathMySQL  BootPath = "MySQL"
	BootPathLegacy BootPath = "Legacy"
)

type HAAction string

const (
	ActionBoot HAAction = "BOOT"
	ActionSkip HAAction = "SKIP"
	ActionWait HAAction = "WAIT"
)

type HADecision struct {
	VMID       string
	Action     HAAction
	Path       BootPath
	TargetNode string
	Priority   int
	Reason     string
}

type BootTaskStatus string

const (
	TaskPending   BootTaskStatus = "PENDING"
	TaskRunning   BootTaskStatus = "RUNNING"
	TaskCompleted BootTaskStatus = "COMPLETED"
	TaskFailed    BootTaskStatus = "FAILED"
)

type BootTask struct {
	VMID       string
	Decision   HADecision
	Status     BootTaskStatus
	RetryCount int
	LastError  error
}

type HAEngineConfig struct {
	MaxConcurrentBoots int
	BootTimeoutMs      int
	MaxRetries         int
}

type ClusterView struct {
	Nodes map[string]fdm.NodeState
}

type BatchBootPolicy struct {
	MaxConcurrent int
}

//Personal.AI order the ending