package heartbeat

import (
	"context"
	"time"

	"github.com/turtacn/hci-vcls/pkg/witness"
)

type heartbeaterImpl struct {
	config       HeartbeatConfig
	peerStates   map[string]HeartbeatState
	deadCbs      []func(string)
	recoveredCbs []func(string)
	witnessPool  witness.Pool
	ctx          context.Context
	cancel       context.CancelFunc
}

func NewHeartbeater(config HeartbeatConfig, pool witness.Pool) Heartbeater {
	ctx, cancel := context.WithCancel(context.Background())
	hb := &heartbeaterImpl{
		config:       config,
		peerStates:   make(map[string]HeartbeatState),
		deadCbs:      make([]func(string), 0),
		recoveredCbs: make([]func(string), 0),
		witnessPool:  pool,
		ctx:          ctx,
		cancel:       cancel,
	}

	for _, peer := range config.Peers {
		hb.peerStates[peer] = HeartbeatState{LastSeen: time.Now(), IsAlive: true}
	}

	return hb
}

func (h *heartbeaterImpl) Start(ctx context.Context) error {
	go h.monitorLoop(ctx)
	return nil
}

func (h *heartbeaterImpl) Stop() error {
	h.cancel()
	return nil
}

func (h *heartbeaterImpl) PeerState(nodeID string) (HeartbeatState, error) {
	state, ok := h.peerStates[nodeID]
	if !ok {
		return HeartbeatState{}, ErrPeerNotFound
	}
	return state, nil
}

func (h *heartbeaterImpl) AllPeerStates() map[string]HeartbeatState {
	return h.peerStates
}

func (h *heartbeaterImpl) OnPeerDead(callback func(nodeID string)) {
	h.deadCbs = append(h.deadCbs, callback)
}

func (h *heartbeaterImpl) OnPeerRecovered(callback func(nodeID string)) {
	h.recoveredCbs = append(h.recoveredCbs, callback)
}

func (h *heartbeaterImpl) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(h.config.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			for peerID, state := range h.peerStates {
				if time.Since(state.LastSeen) > time.Duration(h.config.TimeoutMs)*time.Millisecond {
					if state.IsAlive {
						state.IsAlive = false
						h.peerStates[peerID] = state
						h.notifyDead(peerID)
					}
				}
			}
		}
	}
}

func (h *heartbeaterImpl) notifyDead(nodeID string) {
	for _, cb := range h.deadCbs {
		cb(nodeID)
	}
}

// func (h *heartbeaterImpl) notifyRecovered(nodeID string) {
// 	for _, cb := range h.recoveredCbs {
// 		cb(nodeID)
// 	}
// }

