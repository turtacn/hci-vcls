package mysql

import "time"

type HealthState int

const (
	MySQLStateHealthy HealthState = iota
	MySQLStateReadOnly
	MySQLStateUnavailable
)

func (s HealthState) String() string {
	switch s {
	case MySQLStateHealthy:
		return "Healthy"
	case MySQLStateReadOnly:
		return "ReadOnly"
	case MySQLStateUnavailable:
		return "Unavailable"
	default:
		return "Unknown"
	}
}

type MySQLConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type HAVMState struct {
	VMID       string
	Token      string
	TargetNode string
	Status     string
}

type BootClaim struct {
	VMID       string
	Token      string
	TargetNode string
	UpdatedAt  time.Time
}

type MySQLStatus struct {
	State HealthState
	Error error
}

type VMRecord struct {
	VMID        string
	ClusterID   string
	CurrentHost string
	PowerState  string
	Protected   bool
	UpdatedAt   time.Time
}

type FDMRecord struct {
	ClusterID        string
	Degradation      float64
	HeartbeatLossSum int
	Reason           string
	ComputedAt       time.Time
}

type WitnessRecord struct {
	VMID      string
	ClusterID string
	Available bool
	CheckedAt time.Time
}

type TaskStatus string

const (
	TaskPending   TaskStatus = "PENDING"
	TaskRunning   TaskStatus = "RUNNING"
	TaskCompleted TaskStatus = "COMPLETED"
	TaskFailed    TaskStatus = "FAILED"
)

type HATaskRecord struct {
	ID         string
	VMID       string
	ClusterID  string
	SourceHost string
	TargetHost string
	BootPath   string
	Status     TaskStatus
	BatchNo    int
	RetryCount int
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PlanRecord struct {
	ID          string
	ClusterID   string
	Trigger     string
	Degradation float64
	TaskCount   int
	CreatedAt   time.Time
}
