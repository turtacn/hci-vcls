package config

import "errors"

var (
	ErrConfigFileNotFound = errors.New("config file not found")
	ErrInvalidConfig      = errors.New("invalid configuration")
	ErrDecoding           = errors.New("unable to decode configuration")
)

//Personal.AI order the ending
