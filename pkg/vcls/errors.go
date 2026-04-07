package vcls

import "fmt"

type CapabilityError struct {
	Code    string
	Message string
	Err     error
}

func (e *CapabilityError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("vcls error %s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("vcls error %s: %s", e.Code, e.Message)
}

func (e *CapabilityError) Unwrap() error {
	return e.Err
}

var (
	ErrCapabilityUnavailable   = &CapabilityError{Code: "ERR_CAPABILITY_UNAVAILABLE", Message: "capability is unavailable"}
	ErrAgentNotStarted         = &CapabilityError{Code: "ERR_AGENT_NOT_STARTED", Message: "vcls agent not started"}
	ErrInvalidCapabilityMatrix = &CapabilityError{Code: "ERR_INVALID_CAPABILITY_MATRIX", Message: "invalid capability matrix"}
)

