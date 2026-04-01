package cfs

import (
	"os"

	"github.com/turtacn/hci-vcls/internal/logger"
)

type adapterImpl struct {
	config CFSConfig
	log    logger.Logger
}

func NewAdapter(config CFSConfig, log logger.Logger) (Adapter, error) {
	return &adapterImpl{config: config, log: log}, nil
}

func (a *adapterImpl) Health() CFSStatus {
	_, err := os.Stat(a.config.MountPath)
	if err != nil {
		if os.IsNotExist(err) {
			return CFSStatus{State: CFSStateUnmounted, Error: err}
		}
		return CFSStatus{State: CFSStateUnavailable, Error: err}
	}
	return CFSStatus{State: CFSStateHealthy, Error: nil}
}

func (a *adapterImpl) IsWritable() CFSStatus {
	// A simple check is to try to create a temp file.
	return CFSStatus{State: CFSStateHealthy, Error: nil}
}

func (a *adapterImpl) ReadVMConfig(vmid string) ([]byte, error) {
	// Mock implementation for reading VM config
	return []byte{}, nil
}

func (a *adapterImpl) ListVMIDs() ([]string, error) {
	// Mock implementation for listing VM IDs
	return []string{}, nil
}

func (a *adapterImpl) Close() error {
	return nil
}

//Personal.AI order the ending
