package mysql

import (
	"errors"
	"strings"
	"testing"
)

func TestErrors(t *testing.T) {
	errs := []error{
		ErrBootTokenConflict,
		ErrBootTokenMismatch,
		ErrVMStateNotFound,
		ErrMySQLReadOnly,
		ErrConnectionPool,
		ErrOptimisticLockFailed,
	}

	for _, err := range errs {
		if len(err.Error()) == 0 {
			t.Errorf("Error %v has no message", err)
		}
		if !strings.HasPrefix(err.Error(), "mysql error") {
			t.Errorf("Error %v missing prefix", err)
		}
	}
}

func TestMySQLError_Unwrap(t *testing.T) {
	baseErr := errors.New("base")
	myErr := &MySQLError{Code: "C", Message: "M", Err: baseErr}
	if myErr.Unwrap() != baseErr {
		t.Errorf("expected to unwrap base error")
	}
	if !strings.Contains(myErr.Error(), "base") {
		t.Errorf("expected string to contain inner error text")
	}
}
