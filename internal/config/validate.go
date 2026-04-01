package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
)

func (c *Config) Validate() error {
	validate := validator.New()
	if err := validate.Struct(c); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidConfig, err)
	}
	return nil
}

//Personal.AI order the ending
