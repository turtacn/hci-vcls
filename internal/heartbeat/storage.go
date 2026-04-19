package heartbeat

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/turtacn/hci-vcls/internal/logger"
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
	log          logger.Logger
}

var _ Heartbeater = &StorageHeartbeater{}

func NewStorageHeartbeater(config HeartbeatConfig, dir string, log logger.Logger) *StorageHeartbeater {
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
		log: log,
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

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			h.digestMu.Lock()
			h.stateDigest.Timestamp = time.Now()
			timestamp := h.stateDigest.Timestamp
			h.digestMu.Unlock()

			h.Write(h.config.NodeID, timestamp)
		}
	}
}

func (h *StorageHeartbeater) Write(nodeID string, timestamp time.Time) {
	nodeDir := filepath.Join(h.dir, nodeID)
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		if h.log != nil {
			h.log.Warn("Failed to create storage heartbeat directory", "dir", nodeDir, "error", err)
		}
		return
	}

	filename := filepath.Join(nodeDir, "hb.json")
	tmpFilename := filename + ".tmp"

	h.digestMu.Lock()
	h.stateDigest.Timestamp = timestamp
	data, _ := json.Marshal(h.stateDigest)
	h.digestMu.Unlock()

	// Atomic write
	if err := os.WriteFile(tmpFilename, data, 0644); err == nil {
		if err := os.Rename(tmpFilename, filename); err != nil {
			if h.log != nil {
				h.log.Warn("Failed to rename storage heartbeat file", "node", nodeID, "error", err)
			}
		}
	} else {
		if h.log != nil {
			h.log.Warn("Failed to write to storage heartbeat file", "node", nodeID, "error", err)
		}
	}
}

func (h *StorageHeartbeater) Read(nodeID string) (time.Time, error) {
	filename := filepath.Join(h.dir, nodeID, "hb.json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return time.Time{}, err
	}

	var digest StateDigest
	if err := json.Unmarshal(data, &digest); err != nil {
		return time.Time{}, err
	}

	return digest.Timestamp, nil
}

func (h *StorageHeartbeater) ReadAll() map[string]time.Time {
	res := make(map[string]time.Time)
	files, err := os.ReadDir(h.dir)
	if err != nil {
		if h.log != nil {
			h.log.Warn("Failed to read storage heartbeat directory", "dir", h.dir, "error", err)
		}
		return res
	}

	for _, f := range files {
		if f.IsDir() {
			nodeID := f.Name()
			ts, err := h.Read(nodeID)
			if err == nil {
				res[nodeID] = ts
			}
		}
	}

	return res
}

func (h *StorageHeartbeater) readLoop() {
	ticker := time.NewTicker(time.Duration(h.config.IntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			allTimes := h.ReadAll()

			h.peerStatesMu.Lock()
			for peer := range h.peerStates {
				if peer == h.config.NodeID {
					continue
				}

				ts, ok := allTimes[peer]
				if !ok {
					continue
				}

				state := h.peerStates[peer]
				if ts.After(state.LastSeen) {
					wasAlive := state.IsAlive
					state.LastSeen = ts
					state.IsAlive = true
					h.peerStates[peer] = state

					if !wasAlive {
						h.notifyRecovered(peer)
					}

					filename := filepath.Join(h.dir, peer, "hb.json")
					if data, err := os.ReadFile(filename); err == nil {
						var digest StateDigest
						if json.Unmarshal(data, &digest) == nil {
							for _, cb := range h.digestCbs {
								cb(digest)
							}
						}
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

