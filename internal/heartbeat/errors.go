package heartbeat

import "fmt"

type HeartbeatError struct {
	Code    string
	Message string
	Err     error
}

func (e *HeartbeatError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("heartbeat error %s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("heartbeat error %s: %s", e.Code, e.Message)
}

func (e *HeartbeatError) Unwrap() error {
	return e.Err
}

var (
	ErrPeerNotFound     = &HeartbeatError{Code: "ERR_PEER_NOT_FOUND", Message: "peer not found"}
	ErrHeartbeatStopped = &HeartbeatError{Code: "ERR_HEARTBEAT_STOPPED", Message: "heartbeat is stopped"}
	ErrHeartbeatTimeout = &HeartbeatError{Code: "ERR_HEARTBEAT_TIMEOUT", Message: "heartbeat timeout"}
)

//Personal.AI order the ending
