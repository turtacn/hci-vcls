package qm

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

//Personal.AI order the ending
