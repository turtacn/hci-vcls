package config

import "time"

type NodeConfig struct {
	NodeID    string `mapstructure:"node_id" validate:"required"`
	ClusterID string `mapstructure:"cluster_id" validate:"required"`
	HostIP    string `mapstructure:"host_ip" validate:"required"`
}

type ServerConfig struct {
	HTTPAddr string `mapstructure:"http_addr" validate:"required"`
	GRPCAddr string `mapstructure:"grpc_addr" validate:"required"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type ZKConfig struct {
	Endpoints      []string      `mapstructure:"endpoints" validate:"required,min=1"`
	SessionTimeout time.Duration `mapstructure:"session_timeout" validate:"required"`
}

type MySQLConfig struct {
	DSN          string `mapstructure:"dsn" validate:"required"`
	MaxOpenConns int    `mapstructure:"max_open_conns" validate:"min=1"`
}

type HeartbeatConfig struct {
	Interval   time.Duration `mapstructure:"interval" validate:"required,gt=0"`
	Timeout    time.Duration `mapstructure:"timeout" validate:"required,gtfield=Interval"`
	WindowSize int           `mapstructure:"window_size" validate:"min=1"`
}

type ElectionConfig struct {
	LeaseTTL      time.Duration `mapstructure:"lease_ttl" validate:"required,gtfield=RenewInterval"`
	RenewInterval time.Duration `mapstructure:"renew_interval" validate:"required,gt=0"`
}

type FDMConfig struct {
	EvalInterval      time.Duration `mapstructure:"eval_interval" validate:"required,gt=0"`
	QuorumRatio       float64       `mapstructure:"quorum_ratio" validate:"gt=0,lte=1"`
	CriticalThreshold int           `mapstructure:"critical_threshold"`
}

type HAConfig struct {
	BatchSize     int           `mapstructure:"batch_size" validate:"required,gt=0"`
	BatchInterval time.Duration `mapstructure:"batch_interval" validate:"gt=0"`
	Cooldown      time.Duration `mapstructure:"cooldown"`
	FailFast      bool          `mapstructure:"fail_fast"`
}

type CacheConfig struct {
	DefaultTTL time.Duration `mapstructure:"default_ttl" validate:"required"`
	MaxEntries int           `mapstructure:"max_entries" validate:"min=1"`
}

type MetricsConfig struct {
	Addr      string `mapstructure:"addr" validate:"required"`
	Namespace string `mapstructure:"namespace"`
}

type GRPCConfig struct {
	MaxRecvMsgSize int `mapstructure:"max_recv_msg_size" validate:"min=1"`
}

type Config struct {
	Node      NodeConfig      `mapstructure:"node"`
	Server    ServerConfig    `mapstructure:"server"`
	Log       LogConfig       `mapstructure:"log"`
	ZK        ZKConfig        `mapstructure:"zk"`
	MySQL     MySQLConfig     `mapstructure:"mysql"`
	Heartbeat HeartbeatConfig `mapstructure:"heartbeat"`
	Election  ElectionConfig  `mapstructure:"election"`
	FDM       FDMConfig       `mapstructure:"fdm"`
	HA        HAConfig        `mapstructure:"ha"`
	Cache     CacheConfig     `mapstructure:"cache"`
	Metrics   MetricsConfig   `mapstructure:"metrics"`
	GRPC      GRPCConfig      `mapstructure:"grpc"`
}

