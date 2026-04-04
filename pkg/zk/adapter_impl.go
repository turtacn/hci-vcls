package zk

import (
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/turtacn/hci-vcls/internal/logger"
)

type adapterImpl struct {
	conn *zk.Conn
	log  logger.Logger
}

func NewAdapter(config ZKConfig, log logger.Logger) (Adapter, error) {
	conn, _, err := zk.Connect(config.Endpoints, time.Duration(config.SessionTimeoutMs)*time.Millisecond)
	if err != nil {
		return nil, err
	}
	return &adapterImpl{conn: conn, log: log}, nil
}

func (a *adapterImpl) Health() ZKStatus {
	state := a.conn.State()
	if state == zk.StateHasSession {
		return ZKStatus{State: ZKStateHealthy, Error: nil}
	}
	return ZKStatus{State: ZKStateUnavailable, Error: nil}
}

func (a *adapterImpl) IsReadOnly() ZKStatus {
	// A real implementation would check the "isro" 4-letter word or connection metadata.
	return ZKStatus{State: ZKStateHealthy, Error: nil}
}

func (a *adapterImpl) Ping() ZKStatus {
	_, _, err := a.conn.Exists("/")
	if err != nil {
		return ZKStatus{State: ZKStateUnavailable, Error: err}
	}
	return ZKStatus{State: ZKStateHealthy, Error: nil}
}

func (a *adapterImpl) Close() error {
	if a.conn != nil {
		a.conn.Close()
	}
	return nil
}

// Personal.AI order the ending
