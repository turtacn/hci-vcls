package heartbeat

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type StorageHeartbeater struct {
	config       HeartbeatConfig
	dir          string
	peerStates   map[string]HeartbeatState
	peerStatesMu sync.RWMutex
	deadCbs      []func(string)
	recoveredCbs []func(string)
	digestCbs    []func(StateDigest)
	ctx          context.Context
	cancel       context.CancelFunc
	stateDigest  *StateDigest
	digestMu     sync.RWMutex
}

var _ Heartbeater = &StorageHeartbeater{}

func NewStorageHeartbeater(config HeartbeatConfig, dir string) *StorageHeartbeater {
	ctx, cancel := context.WithCancel(context.Background())
	hb := &StorageHeartbeater{
		config:       config,
		dir:          dir,
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

func (h *StorageHeartbeater) Start(ctx context.Context) error {
	if err := os.MkdirAll(h.dir, 0755); err != nil {
		return err
	}

	go h.writeLoop()
	go h.readLoop()
	go h.monitorLoop(h.ctx)

	return nil
}

func (h *StorageHeartbeater) Stop() error {
	h.cancel()
	return nil
}

func (h *StorageHeartbeater) PeerState(nodeID string) (HeartbeatState, error) {
	h.peerStatesMu.RLock()
	defer h.peerStatesMu.RUnlock()
	state, ok := h.peerStates[nodeID]
	if !ok {
		return HeartbeatState{}, ErrPeerNotFound
	}
	return state, nil
}

func (h *StorageHeartbeater) AllPeerStates() map[string]HeartbeatState {
	h.peerStatesMu.RLock()
	defer h.peerStatesMu.RUnlock()

	res := make(map[string]HeartbeatState)
	for k, v := range h.peerStates {
		res[k] = v
	}
	return res
}

func (h *StorageHeartbeater) OnPeerDead(callback func(nodeID string)) {
	h.deadCbs = append(h.deadCbs, callback)
}

func (h *StorageHeartbeater) OnPeerRecovered(callback func(nodeID string)) {
	h.recoveredCbs = append(h.recoveredCbs, callback)
}

func (h *StorageHeartbeater) OnDigestReceived(callback func(digest StateDigest)) {
	h.digestCbs = append(h.digestCbs, callback)
}

func (h *StorageHeartbeater) UpdateDigest(term int64, candidateID string, isLeader bool) {
	h.digestMu.Lock()
	defer h.digestMu.Unlock()
	h.stateDigest.Term = term
	h.stateDigest.CandidateID = candidateID
	h.stateDigest.IsLeader = isLeader
}

func (h *StorageHeartbeater) writeLoop() {
	ticker := time.NewTicker(time.Duration(h.config.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	filename := filepath.Join(h.dir, h.config.NodeID+".hb")
	tmpFilename := filename + ".tmp"

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.digestMu.RLock()
			h.stateDigest.Timestamp = time.Now()
			data, _ := json.Marshal(h.stateDigest)
			h.digestMu.RUnlock()

			// Atomic write
			if err := os.WriteFile(tmpFilename, data, 0644); err == nil {
				os.Rename(tmpFilename, filename)
			}
		}
	}
}

func (h *StorageHeartbeater) readLoop() {
	ticker := time.NewTicker(time.Duration(h.config.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.peerStatesMu.Lock()
			for peer := range h.peerStates {
				if peer == h.config.NodeID {
					continue
				}

				filename := filepath.Join(h.dir, peer+".hb")
				data, err := os.ReadFile(filename)
				if err != nil {
					continue
				}

				var digest StateDigest
				if err := json.Unmarshal(data, &digest); err != nil {
					continue
				}

				state := h.peerStates[peer]
				// Basic check to see if timestamp updated
				if digest.Timestamp.After(state.LastSeen) {
					wasAlive := state.IsAlive
					state.LastSeen = digest.Timestamp
					state.IsAlive = true
					h.peerStates[peer] = state

					if !wasAlive {
						h.notifyRecovered(peer)
					}

					// Only trigger digest received if it's a new update
					for _, cb := range h.digestCbs {
						cb(digest)
					}
				}
			}
			h.peerStatesMu.Unlock()
		}
	}
}

func (h *StorageHeartbeater) monitorLoop(ctx context.Context) {
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

func (h *StorageHeartbeater) notifyDead(nodeID string) {
	for _, cb := range h.deadCbs {
		cb(nodeID)
	}
}

func (h *StorageHeartbeater) notifyRecovered(nodeID string) {
	for _, cb := range h.recoveredCbs {
		cb(nodeID)
	}
}

