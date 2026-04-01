package ha

import "fmt"

type BootError struct {
	Code    string
	Message string
	Err     error
}

func (e *BootError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("ha error %s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("ha error %s: %s", e.Code, e.Message)
}

func (e *BootError) Unwrap() error {
	return e.Err
}

var (
	ErrNotLeader             = &BootError{Code: "ERR_NOT_LEADER", Message: "not the leader node"}
	ErrBootTimeout           = &BootError{Code: "ERR_BOOT_TIMEOUT", Message: "boot task timed out"}
	ErrMaxRetriesExceeded    = &BootError{Code: "ERR_MAX_RETRIES_EXCEEDED", Message: "max boot retries exceeded"}
	ErrNoQuorum              = &BootError{Code: "ERR_NO_QUORUM", Message: "no quorum reached for boot decision"}
	ErrBootTokenLost         = &BootError{Code: "ERR_BOOT_TOKEN_LOST", Message: "boot token lost during operation"}
	ErrVMAlreadyRunning      = &BootError{Code: "ERR_VM_ALREADY_RUNNING", Message: "vm is already running on another node"}
	ErrInsufficientResources = &BootError{Code: "ERR_INSUFFICIENT_RESOURCES", Message: "insufficient resources on target node"}
)

//Personal.AI order the ending
