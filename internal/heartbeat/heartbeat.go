package heartbeat

import "context"

type Heartbeater interface {
	Start(ctx context.Context) error
	Stop() error
	PeerState(nodeID string) (HeartbeatState, error)
	AllPeerStates() map[string]HeartbeatState
	OnPeerDead(callback func(nodeID string))
	OnPeerRecovered(callback func(nodeID string))
}

//Personal.AI order the ending