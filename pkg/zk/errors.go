package zk

import "errors"

var (
	ErrNotConnected    = errors.New("zk: client not connected")
	ErrNodeNotFound    = errors.New("zk: node does not exist")
	ErrNodeExists      = errors.New("zk: node already exists")
	ErrVersionMismatch = errors.New("zk: bad version")
)

