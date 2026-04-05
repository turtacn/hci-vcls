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

func (a *adapterImpl) BeginTx() (TxAdapter, error) {
	tx, err := a.db.Begin()
	if err != nil {
		return nil, err
	}
	return &txAdapterImpl{tx: tx, log: a.log}, nil
}

type txAdapterImpl struct {
	tx  *sql.Tx
	log logger.Logger
}

func (t *txAdapterImpl) ClaimBoot(claim BootClaim) error {
	// 乐观锁 + 状态检查
	res, err := t.tx.Exec(`UPDATE ha_vm_state
		SET status='booting', target_node=?, token=?
		WHERE vmid=? AND status IN ('stopped', 'failed')`,
		claim.TargetNode, claim.Token, claim.VMID)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrOptimisticLockFailed
	}

	return nil
}

func (t *txAdapterImpl) Commit() error {
	return t.tx.Commit()
}

func (t *txAdapterImpl) Rollback() error {
	return t.tx.Rollback()
}

func (a *adapterImpl) ClaimBoot(claim BootClaim) error {
	res, err := a.db.Exec(`UPDATE ha_vm_state
		SET status='booting', target_node=?, token=?
		WHERE vmid=? AND status IN ('stopped', 'failed')`,
		claim.TargetNode, claim.Token, claim.VMID)

	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrOptimisticLockFailed
	}

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
