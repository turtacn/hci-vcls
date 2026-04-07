package fdm

import "fmt"

type FDMError struct {
	Code    string
	Message string
	Err     error
}

func (e *FDMError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("fdm error %s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("fdm error %s: %s", e.Code, e.Message)
}

func (e *FDMError) Unwrap() error {
	return e.Err
}

var (
	ErrNodeNotFound         = &FDMError{Code: "ERR_NODE_NOT_FOUND", Message: "node not found"}
	ErrQuorumNotReached     = &FDMError{Code: "ERR_QUORUM_NOT_REACHED", Message: "quorum not reached"}
	ErrAgentNotStarted      = &FDMError{Code: "ERR_AGENT_NOT_STARTED", Message: "agent not started"}
	ErrDaemonNotStarted     = &FDMError{Code: "ERR_DAEMON_NOT_STARTED", Message: "daemon not started"}
	ErrHeartbeatTimeout     = &FDMError{Code: "ERR_HEARTBEAT_TIMEOUT", Message: "heartbeat timeout"}
	ErrLeaderElectionFailed = &FDMError{Code: "ERR_LEADER_ELECTION_FAILED", Message: "leader election failed"}
	ErrSelfIsolated         = &FDMError{Code: "ERR_SELF_ISOLATED", Message: "self isolated"}
)

