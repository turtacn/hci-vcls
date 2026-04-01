package cfs

type HealthState int

const (
	CFSStateHealthy HealthState = iota
	CFSStateReadOnly
	CFSStateUnmounted
	CFSStateUnavailable
)

func (s HealthState) String() string {
	switch s {
	case CFSStateHealthy:
		return "Healthy"
	case CFSStateReadOnly:
		return "ReadOnly"
	case CFSStateUnmounted:
		return "Unmounted"
	case CFSStateUnavailable:
		return "Unavailable"
	default:
		return "Unknown"
	}
}

type CFSConfig struct {
	MountPath string
	TimeoutMs int
}

type CFSStatus struct {
	State HealthState
	Error error
}

//Personal.AI order the ending
