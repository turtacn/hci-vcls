package election

import "context"

type Elector interface {
	Campaign(ctx context.Context) error
	Resign(ctx context.Context) error
	IsLeader() bool
	LeaderID() (string, error)
	OnLeaderChange(callback func(LeaderInfo))
	Close() error
}

//Personal.AI order the ending