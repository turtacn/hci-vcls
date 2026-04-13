package mysql

import (
	"context"
	"time"
)

type Adapter interface {
	Health() MySQLStatus
	BeginTx() (TxAdapter, error)
	ClaimBoot(claim BootClaim) error
	ConfirmBoot(vmid, token string) error
	ReleaseBoot(vmid, token string) error
	GetVMState(vmid string) (*HAVMState, error)
	UpsertVMState(state HAVMState) error
	Close() error
	ListStaleBootingClaims(ctx context.Context, threshold time.Time) ([]BootClaim, error)
	ReleaseStaleClaim(ctx context.Context, vmid, token string, reason string) error
}

type TxAdapter interface {
	ClaimBoot(claim BootClaim) error
	Commit() error
	Rollback() error
}
