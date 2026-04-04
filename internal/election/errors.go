package election

import "fmt"

type ElectionError struct {
	Code    string
	Message string
	Err     error
}

func (e *ElectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("election error %s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("election error %s: %s", e.Code, e.Message)
}

func (e *ElectionError) Unwrap() error {
	return e.Err
}

var (
	ErrNoLeader            = &ElectionError{Code: "ERR_NO_LEADER", Message: "no leader elected"}
	ErrElectionClosed      = &ElectionError{Code: "ERR_ELECTION_CLOSED", Message: "election is closed"}
	ErrInvalidElectionPath = &ElectionError{Code: "ERR_INVALID_ELECTION_PATH", Message: "invalid election path"}
)

// Personal.AI order the ending
