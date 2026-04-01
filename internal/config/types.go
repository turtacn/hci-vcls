package config

import (
	"github.com/turtacn/hci-vcls/internal/election"
	"github.com/turtacn/hci-vcls/internal/heartbeat"
	"github.com/turtacn/hci-vcls/pkg/cache"
	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/ha"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/qm"
	"github.com/turtacn/hci-vcls/pkg/statemachine"
	"github.com/turtacn/hci-vcls/pkg/vcls"
	"github.com/turtacn/hci-vcls/pkg/witness"
	"github.com/turtacn/hci-vcls/pkg/zk"
)

type NodeConfig struct {
	ID        string `mapstructure:"ID" validate:"required"`
	ClusterID string `mapstructure:"ClusterID" validate:"required"`
}

type ServerConfig struct {
	ListenAddr string `mapstructure:"ListenAddr" validate:"required,hostname_port"`
}

type MetricsConfig struct {
	Enabled    bool   `mapstructure:"Enabled"`
	ListenAddr string `mapstructure:"ListenAddr" validate:"required_if=Enabled true"`
}

type GRPCConfig struct {
	Enabled    bool   `mapstructure:"Enabled"`
	ListenAddr string `mapstructure:"ListenAddr" validate:"required_if=Enabled true"`
}

type Config struct {
	Node         NodeConfig                      `mapstructure:"Node" validate:"required"`
	Server       ServerConfig                    `mapstructure:"Server" validate:"required"`
	Metrics      MetricsConfig                   `mapstructure:"Metrics"`
	GRPC         GRPCConfig                      `mapstructure:"GRPC"`
	ZK           zk.ZKConfig                     `mapstructure:"ZK" validate:"required"`
	CFS          cfs.CFSConfig                   `mapstructure:"CFS" validate:"required"`
	MySQL        mysql.MySQLConfig               `mapstructure:"MySQL" validate:"required"`
	QM           qm.QMConfig                     `mapstructure:"QM" validate:"required"`
	Witness      witness.WitnessConfig           `mapstructure:"Witness"`
	Daemon       fdm.DaemonConfig                `mapstructure:"Daemon" validate:"required"`
	Cache        cache.CacheManagerConfig        `mapstructure:"Cache" validate:"required"`
	HA           ha.HAEngineConfig               `mapstructure:"HA" validate:"required"`
	VCLS         vcls.VCLSConfig                 `mapstructure:"VCLS" validate:"required"`
	StateMachine statemachine.StateMachineConfig `mapstructure:"StateMachine" validate:"required"`
	Election     election.ElectionConfig         `mapstructure:"Election" validate:"required"`
	Heartbeat    heartbeat.HeartbeatConfig       `mapstructure:"Heartbeat" validate:"required"`
	LogLevel     string                          `mapstructure:"LogLevel" validate:"oneof=debug info warn error"`
	LogFormat    string                          `mapstructure:"LogFormat" validate:"oneof=text json"`
}

//Personal.AI order the ending
