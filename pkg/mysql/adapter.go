package mysql

type Adapter interface {
	Health() MySQLStatus
	ClaimBoot(claim BootClaim) error
	ConfirmBoot(vmid, token string) error
	ReleaseBoot(vmid, token string) error
	GetVMState(vmid string) (*HAVMState, error)
	UpsertVMState(state HAVMState) error
	Close() error
}

//Personal.AI order the ending