package cfs

type Adapter interface {
	Health() CFSStatus
	IsWritable() CFSStatus
	ReadVMConfig(vmid string) ([]byte, error)
	ListVMIDs() ([]string, error)
	Close() error
}

