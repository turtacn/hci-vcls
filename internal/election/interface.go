package election

import "context"

type Elector interface {
	Campaign(ctx context.Context) error
	Resign(ctx context.Context) error
	IsLeader() bool
	Status() LeaderStatus
	Watch() <-chan LeaderStatus
	Close() error
	OnLeaderChange(callback func(info LeaderInfo)) // backward compat
}

