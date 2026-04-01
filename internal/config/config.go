package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvPrefix("VCLS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Set defaults
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		// Viper sometimes returns a PathError or similar when the file doesn't exist,
		// depending on how it was initialized. We want to ignore "not found" errors.
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Also check if it's a generic file not found error using errors.Is
			if !errors.Is(err, os.ErrNotExist) {
				var pathErr *os.PathError
				if !errors.As(err, &pathErr) {
					return nil, fmt.Errorf("error reading config file: %w", err)
				}
			}
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecoding, err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
} //Personal.AI order the ending
