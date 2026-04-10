package heartbeat

import (
	"context"
	"encoding/json"
	"net"
	"sync"
	"time"
)

type UDPHeartbeater struct {
	config       HeartbeatConfig
	peerStates   map[string]HeartbeatState
	peerStatesMu sync.RWMutex
	deadCbs      []func(string)
	recoveredCbs []func(string)
	digestCbs    []func(StateDigest)
	ctx          context.Context
	cancel       context.CancelFunc
	conn         *net.UDPConn
	stateDigest  *StateDigest
	digestMu     sync.RWMutex
}

var _ Heartbeater = &UDPHeartbeater{}

func NewUDPHeartbeater(config HeartbeatConfig) *UDPHeartbeater {
	ctx, cancel := context.WithCancel(context.Background())
	hb := &UDPHeartbeater{
		config:       config,
		peerStates:   make(map[string]HeartbeatState),
		deadCbs:      make([]func(string), 0),
		recoveredCbs: make([]func(string), 0),
		digestCbs:    make([]func(StateDigest), 0),
		ctx:          ctx,
		cancel:       cancel,
		stateDigest: &StateDigest{
			NodeID: config.NodeID,
		},
	}

	for _, peer := range config.Peers {
		hb.peerStates[peer] = HeartbeatState{LastSeen: time.Now(), IsAlive: true}
	}

	return hb
}

func (h *UDPHeartbeater) Start(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", h.config.NodeID) // Listen on own addr
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	h.conn = conn

	go h.receiveLoop()
	go h.sendLoop()
	go h.monitorLoop(h.ctx)

	return nil
}

func (h *UDPHeartbeater) Stop() error {
	h.cancel()
	if h.conn != nil {
		return h.conn.Close()
	}
	return nil
}

func (h *UDPHeartbeater) PeerState(nodeID string) (HeartbeatState, error) {
	h.peerStatesMu.RLock()
	defer h.peerStatesMu.RUnlock()
	state, ok := h.peerStates[nodeID]
	if !ok {
		return HeartbeatState{}, ErrPeerNotFound
	}
	return state, nil
}

func (h *UDPHeartbeater) AllPeerStates() map[string]HeartbeatState {
	h.peerStatesMu.RLock()
	defer h.peerStatesMu.RUnlock()

	res := make(map[string]HeartbeatState)
	for k, v := range h.peerStates {
		res[k] = v
	}
	return res
}

func (h *UDPHeartbeater) OnPeerDead(callback func(nodeID string)) {
	h.deadCbs = append(h.deadCbs, callback)
}

func (h *UDPHeartbeater) OnPeerRecovered(callback func(nodeID string)) {
	h.recoveredCbs = append(h.recoveredCbs, callback)
}

func (h *UDPHeartbeater) OnDigestReceived(callback func(digest StateDigest)) {
	h.digestCbs = append(h.digestCbs, callback)
}

func (h *UDPHeartbeater) UpdateDigest(term int64, candidateID string, isLeader bool) {
	h.digestMu.Lock()
	defer h.digestMu.Unlock()
	h.stateDigest.Term = term
	h.stateDigest.CandidateID = candidateID
	h.stateDigest.IsLeader = isLeader
}

func (h *UDPHeartbeater) receiveLoop() {
	buf := make([]byte, 1024)
	for {
		select {
		case <-h.ctx.Done():
			return
		default:
		}

		if h.conn == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		h.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, _, err := h.conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			continue
		}

		var digest StateDigest
		if err := json.Unmarshal(buf[:n], &digest); err != nil {
			continue
		}

		h.peerStatesMu.Lock()
		state, ok := h.peerStates[digest.NodeID]
		if ok {
			wasAlive := state.IsAlive
			state.LastSeen = time.Now()
			state.IsAlive = true
			h.peerStates[digest.NodeID] = state
			if !wasAlive {
				h.notifyRecovered(digest.NodeID)
			}
		}
		h.peerStatesMu.Unlock()

		for _, cb := range h.digestCbs {
			cb(digest)
		}
	}
}

func (h *UDPHeartbeater) sendLoop() {
	ticker := time.NewTicker(time.Duration(h.config.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.digestMu.RLock()
			h.stateDigest.Timestamp = time.Now()
			data, _ := json.Marshal(h.stateDigest)
			h.digestMu.RUnlock()

			// Broadcast to all peers
			for _, peerAddr := range h.config.Peers {
				addr, err := net.ResolveUDPAddr("udp", peerAddr)
				if err != nil {
					continue
				}
				if h.conn != nil {
					_, _ = h.conn.WriteToUDP(data, addr)
				}
			}
		}
	}
}

func (h *UDPHeartbeater) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(h.config.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			h.peerStatesMu.Lock()
			for peerID, state := range h.peerStates {
				if time.Since(state.LastSeen) > time.Duration(h.config.TimeoutMs)*time.Millisecond {
					if state.IsAlive {
						state.IsAlive = false
						h.peerStates[peerID] = state
						h.notifyDead(peerID)
					}
				}
			}
			h.peerStatesMu.Unlock()
		}
	}
}

func (h *UDPHeartbeater) notifyDead(nodeID string) {
	for _, cb := range h.deadCbs {
		cb(nodeID)
	}
}

func (h *UDPHeartbeater) notifyRecovered(nodeID string) {
	for _, cb := range h.recoveredCbs {
		cb(nodeID)
	}
}

