package mysql

type HealthState int

const (
	MySQLStateHealthy HealthState = iota
	MySQLStateReadOnly
	MySQLStateUnavailable
)

func (s HealthState) String() string {
	switch s {
	case MySQLStateHealthy:
		return "Healthy"
	case MySQLStateReadOnly:
		return "ReadOnly"
	case MySQLStateUnavailable:
		return "Unavailable"
	default:
		return "Unknown"
	}
}

type MySQLConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type HAVMState struct {
	VMID       string
	Token      string
	TargetNode string
	Status     string
}

type BootClaim struct {
	VMID       string
	Token      string
	TargetNode string
}

type MySQLStatus struct {
	State HealthState
	Error error
}

//Personal.AI order the ending