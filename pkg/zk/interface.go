package zk

import "context"

type Client interface {
	Connect(ctx context.Context) error
	Close() error
	Create(path, data string, ephemeral bool) error
	Get(path string) (string, error)
	Set(path, data string, version int32) error
	Delete(path string, version int32) error
	Exists(path string) (bool, error)
	Watch(path string) (<-chan WatchEvent, error)
	SessionState() SessionState
}

// Personal.AI order the ending
