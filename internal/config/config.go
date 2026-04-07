package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("HCI")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// 1. Set defaults
	setDefaults(v)

	// 2. Read from yaml file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			if !strings.Contains(err.Error(), "no such file or directory") {
				return nil, fmt.Errorf("error reading config file: %w", err)
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unable to decode into struct: %w", err)
	}

	// 3. Validate
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

