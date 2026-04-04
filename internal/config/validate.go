package config

import (
	"fmt"

	validator "github.com/go-playground/validator/v10"
)

func (c *Config) Validate() error {
	validate := validator.New()

	// Register a custom tag for address format if needed, but struct validation is sufficient for most rules here.

	err := validate.Struct(c)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return err
		}

		var errMsgs []string
		for _, err := range err.(validator.ValidationErrors) {
			errMsgs = append(errMsgs, fmt.Sprintf("Field validation for '%s' failed on the '%s' tag", err.Field(), err.Tag()))
		}
		return fmt.Errorf("configuration validation failed: %v", errMsgs)
	}

	return nil
}

// Personal.AI order the ending
