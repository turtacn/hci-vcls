package mysql

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"github.com/turtacn/hci-vcls/internal/logger"
)

type adapterImpl struct {
	db  *sql.DB
	log logger.Logger
}

func NewAdapter(config MySQLConfig, log logger.Logger) (Adapter, error) {
	db, err := sql.Open("mysql", config.DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(config.MaxOpenConns)
	db.SetMaxIdleConns(config.MaxIdleConns)

	return &adapterImpl{db: db, log: log}, nil
}

func (a *adapterImpl) Health() MySQLStatus {
	err := a.db.Ping()
	if err != nil {
		return MySQLStatus{State: MySQLStateUnavailable, Error: err}
	}
	return MySQLStatus{State: MySQLStateHealthy, Error: nil}
}

func (a *adapterImpl) ClaimBoot(claim BootClaim) error {
	// A real implementation would run an UPDATE WHERE condition to grab a lock or claim row.
	return nil
}

func (a *adapterImpl) ConfirmBoot(vmid, token string) error {
	return nil
}

func (a *adapterImpl) ReleaseBoot(vmid, token string) error {
	return nil
}

func (a *adapterImpl) GetVMState(vmid string) (*HAVMState, error) {
	return &HAVMState{VMID: vmid, Token: "mock", Status: "running"}, nil
}

func (a *adapterImpl) UpsertVMState(state HAVMState) error {
	return nil
}

func (a *adapterImpl) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}

// Personal.AI order the ending
