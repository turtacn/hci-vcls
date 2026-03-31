package election

import (
	"context"

	"github.com/turtacn/hci-vcls/pkg/zk"
)

type electorImpl struct {
	config     ElectionConfig
	zkAdapter  zk.Adapter
	isLeader   bool
	leaderID   string
	callbacks  []func(LeaderInfo)
}

func NewElector(config ElectionConfig, adapter zk.Adapter) Elector {
	return &electorImpl{
		config:    config,
		zkAdapter: adapter,
		callbacks: make([]func(LeaderInfo), 0),
	}
}

func (e *electorImpl) Campaign(ctx context.Context) error {
	// A minimal mock campaign using ZK or simulating an election cycle
	e.isLeader = true
	e.leaderID = e.config.NodeID
	e.notifyCallbacks(LeaderInfo{NodeID: e.leaderID, Term: 1})
	return nil
}

func (e *electorImpl) Resign(ctx context.Context) error {
	e.isLeader = false
	e.leaderID = ""
	return nil
}

func (e *electorImpl) IsLeader() bool {
	return e.isLeader
}

func (e *electorImpl) LeaderID() (string, error) {
	if e.leaderID == "" {
		return "", ErrNoLeader
	}
	return e.leaderID, nil
}

func (e *electorImpl) OnLeaderChange(callback func(LeaderInfo)) {
	e.callbacks = append(e.callbacks, callback)
}

func (e *electorImpl) notifyCallbacks(info LeaderInfo) {
	for _, cb := range e.callbacks {
		cb(info)
	}
}

func (e *electorImpl) Close() error {
	return e.Resign(context.Background())
}

//Personal.AI order the ending