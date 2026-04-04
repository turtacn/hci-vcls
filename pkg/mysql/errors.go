package mysql

import "fmt"

type MySQLError struct {
	Code    string
	Message string
	Err     error
}

func (e *MySQLError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("mysql error %s: %s - %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("mysql error %s: %s", e.Code, e.Message)
}

func (e *MySQLError) Unwrap() error {
	return e.Err
}

var (
	ErrBootTokenConflict = &MySQLError{Code: "ERR_BOOT_TOKEN_CONFLICT", Message: "boot token conflict"}
	ErrBootTokenMismatch = &MySQLError{Code: "ERR_BOOT_TOKEN_MISMATCH", Message: "boot token mismatch"}
	ErrVMStateNotFound   = &MySQLError{Code: "ERR_VM_STATE_NOT_FOUND", Message: "vm state not found"}
	ErrMySQLReadOnly     = &MySQLError{Code: "ERR_MYSQL_READ_ONLY", Message: "mysql is read-only"}
	ErrConnectionPool    = &MySQLError{Code: "ERR_CONNECTION_POOL", Message: "mysql connection pool error"}
)

// Personal.AI order the ending
