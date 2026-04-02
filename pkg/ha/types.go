package ha

import (
	"context"
	"time"

	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/vcls"
)

type BootPath string

const (
	BootPathNormal  BootPath = "normal"
	BootPathWitness BootPath = "witness"
)

type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskExecuting TaskStatus = "executing"
	TaskDone      TaskStatus = "done"
	TaskFailed    TaskStatus = "failed"
	TaskSkipped   TaskStatus = "skipped"
)

type HostCandidate struct {
	HostID         string
	Healthy        bool
	CurrentLoad    int // 当前已分配任务数
	FaultDomain    string
	RecentFailures int // 最近 1h 失败次数
	WitnessCapable bool
}

type VMTask struct {
	ID         string
	VMID       string
	ClusterID  string
	SourceHost string
	TargetHost string
	BatchNo    int
	BootPath   BootPath
	Status     TaskStatus
	Score      float64 // 选路分数（可解释）
	Reason     string  // 选路原因，如"Host-B: healthy(+100), low-load(+10), witness-path(+5)"
	RetryCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type Plan struct {
	ID           string
	ClusterID    string
	Trigger      string // "auto" | "manual"
	Degradation  string
	Tasks        []VMTask
	TotalBatches int
	CreatedAt    time.Time
}

type PlanRequest struct {
	ClusterID      string
	FailedHosts    []string
	ProtectedVMs   []*vcls.VM // 来自 vcls.ListEligible
	HostCandidates []HostCandidate
	PreferWitness  bool
	BatchSize      int
}

// Old types to keep tests compiling temporarily
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
	OldTaskPending   BootTaskStatus = "PENDING"
	OldTaskRunning   BootTaskStatus = "RUNNING"
	OldTaskCompleted BootTaskStatus = "COMPLETED"
	OldTaskFailed    BootTaskStatus = "FAILED"
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

const (
	BootPathZK     BootPath = "ZK"
	BootPathCFS    BootPath = "CFS"
	BootPathMySQL  BootPath = "MySQL"
	BootPathLegacy BootPath = "Legacy"
)

type HAEngine interface {
	Start(ctx context.Context) error
	Stop() error
	Evaluate(vmid string) (HADecision, error)
	Execute(decision HADecision) error
	Status(vmid string) BootTaskStatus
	ActiveTasks() []BootTask
}

// Personal.AI order the ending
