// Package config provides a central configuration loader and validator
// for the HCI vCLS application. This package reads configuration settings
// from environment variables and a YAML file (e.g., `config.yaml`),
// applies default values where appropriate, and validates the assembled
// configuration using structural tags. It exposes a single `Config` struct
// holding all parameters for all the application's components.
package config

//Personal.AI order the ending