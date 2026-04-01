package config

import (
	"fmt"
	"strings"

	validator "github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
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

type Config struct {
	ZK           zk.ZKConfig
	CFS          cfs.CFSConfig
	MySQL        mysql.MySQLConfig
	QM           qm.QMConfig
	Witness      witness.WitnessConfig
	Daemon       fdm.DaemonConfig
	Cache        cache.CacheManagerConfig
	HA           ha.HAEngineConfig
	VCLS         vcls.VCLSConfig
	StateMachine statemachine.StateMachineConfig
	Election     election.ElectionConfig
	Heartbeat    heartbeat.HeartbeatConfig
	LogLevel     string
	LogFormat    string
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("VCLS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		// Viper returns different error types/messages depending on how it's initialized.
		// If the file simply doesn't exist, we can ignore the error.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// also check if the error is due to os.ErrNotExist, which can be wrapped
			if !strings.Contains(err.Error(), "no such file or directory") {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("LogLevel", "info")
	v.SetDefault("LogFormat", "text")
	v.SetDefault("ZK.SessionTimeoutMs", 5000)
	v.SetDefault("CFS.TimeoutMs", 3000)
	v.SetDefault("MySQL.MaxOpenConns", 20)
	v.SetDefault("MySQL.MaxIdleConns", 5)
	v.SetDefault("QM.TimeoutMs", 30000)
	v.SetDefault("HA.MaxConcurrentBoots", 5)
	v.SetDefault("HA.BootTimeoutMs", 60000)
	v.SetDefault("HA.MaxRetries", 3)
}

func (c *Config) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

//Personal.AI order the ending