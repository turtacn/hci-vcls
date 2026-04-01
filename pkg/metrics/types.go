package metrics

type MetricLabels map[string]string

const (
	LabelNodeID           = "node_id"
	LabelClusterID        = "cluster_id"
	LabelDegradationLevel = "degradation_level"
	LabelBootPath         = "boot_path"
	LabelVMID             = "vmid"
	LabelResult           = "result"
)

const (
	MetricHABootTotal      = "ha_boot_total"
	MetricHABootDuration   = "ha_boot_duration_seconds"
	MetricFDMHeartbeatLost = "fdm_heartbeat_lost_total"
	MetricDegradationLevel = "degradation_level_info"
	MetricCacheAgeSeconds  = "cache_age_seconds"
	MetricElectionTotal    = "election_total"
	MetricLeaderChanges    = "leader_changes_total"
)

//Personal.AI order the ending
