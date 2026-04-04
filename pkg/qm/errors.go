package qm

import "fmt"

type QMError struct {
	Code    string
	Message string
	Err     error
}

func (e *QMError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("qm error %s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("qm error %s: %s", e.Code, e.Message)
}

func (e *QMError) Unwrap() error {
	return e.Err
}

var (
	ErrVMNotFound       = &QMError{Code: "ERR_VM_NOT_FOUND", Message: "vm not found"}
	ErrVMLocked         = &QMError{Code: "ERR_VM_LOCKED", Message: "vm is locked"}
	ErrVMAlreadyRunning = &QMError{Code: "ERR_VM_ALREADY_RUNNING", Message: "vm is already running"}
	ErrQMTimeout        = &QMError{Code: "ERR_QM_TIMEOUT", Message: "qm operation timed out"}
	ErrSemaphoreTimeout = &QMError{Code: "ERR_SEMAPHORE_TIMEOUT", Message: "semaphore timeout"}
)

// Personal.AI order the ending
