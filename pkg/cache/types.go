package cache

import "time"

type VMComputeMeta struct {
	VMID    string
	CPUs    int
	Memory  int
	NodeID  string
	GroupID string
}

type VMNetworkMeta struct {
	VMID string
	NICs []NICConfig
}

type NICConfig struct {
	MACAddress string
	Network    string
}

type VMStorageMeta struct {
	VMID  string
	Disks []DiskConfig
}

type DiskConfig struct {
	DiskID string
	SizeGB int
	Type   string
}

type VMHAMeta struct {
	VMID    string
	NodeID  string
	State   string
	Token   string
	Retries int
}

type CachedEntry struct {
	Data      interface{}
	Timestamp time.Time
	TTL       time.Duration
}

func (e *CachedEntry) IsExpired() bool {
	if e.TTL <= 0 {
		return false
	}
	return time.Since(e.Timestamp) > e.TTL
}

// Personal.AI order the ending
