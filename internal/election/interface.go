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
	ReceivePeerState(peerNodeID string, peerTerm int64, peerVoteFor string, isLeader bool)
	CurrentTermAndVote() (int64, string, bool)
	SetNodesCount(count int)
}

