package mysql

type Adapter interface {
	Health() MySQLStatus
	BeginTx() (TxAdapter, error)
	ClaimBoot(claim BootClaim) error
	ConfirmBoot(vmid, token string) error
	ReleaseBoot(vmid, token string) error
	GetVMState(vmid string) (*HAVMState, error)
	UpsertVMState(state HAVMState) error
	Close() error
}

type TxAdapter interface {
	ClaimBoot(claim BootClaim) error
	Commit() error
	Rollback() error
}

// Personal.AI order the ending
