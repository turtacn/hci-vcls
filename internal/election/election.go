package election

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type memoryElector struct {
	mu           sync.RWMutex
	nodeID       string
	isLeader     bool
	leaderID     string
	term         int64
	votedFor     string
	votesGranted int
	nodesCount   int
	watcher      *Watcher
	legacyCbs    []func(LeaderInfo)
	termFile     string
	cancel       context.CancelFunc
}

var _ Elector = &memoryElector{}

func NewMemoryElector(nodeID string) *memoryElector {
	termDir := "/var/lib/hci-vcls"
	_ = os.MkdirAll(termDir, 0755)
	termFile := filepath.Join(termDir, "hci-vcls-term-"+nodeID)
	term := loadTerm(termFile)

	return &memoryElector{
		nodeID:     nodeID,
		term:       term,
		watcher:    NewWatcher(),
		legacyCbs:  make([]func(LeaderInfo), 0),
		termFile:   termFile,
		nodesCount: 3, // default to 3 nodes for testing if not set dynamically
	}
}

func (e *memoryElector) SetNodesCount(count int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nodesCount = count
}

func (e *memoryElector) ReceivePeerState(peerNodeID string, peerTerm int64, peerVoteFor string, isLeader bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// If peer term is higher, we step down to follower
	if peerTerm > e.term {
		e.term = peerTerm
		e.isLeader = false
		e.leaderID = ""
		e.votedFor = ""
		e.saveTerm()

		e.notify()
		// NOTE(phase06): term persistence to disk not implemented; see architecture doc §7.6
	}

	// Simple voting logic
	// If a peer is campaigning, it votes for itself (peerVoteFor == peerNodeID)
	if peerVoteFor == peerNodeID && peerTerm >= e.term && (e.votedFor == "" || e.votedFor == peerNodeID) {
		// Vote for the peer
		e.votedFor = peerNodeID
		e.term = peerTerm
		e.saveTerm()
	} else if peerVoteFor == e.nodeID && peerTerm == e.term {
		// They voted for us
		e.votesGranted++
		if !e.isLeader && e.votesGranted >= (e.nodesCount/2)+1 {
			e.isLeader = true
			e.leaderID = e.nodeID
			e.notify()
		}
	}

	// Track current leader if they are explicitly asserting leadership via IsLeader
	if isLeader && peerTerm >= e.term {
		e.leaderID = peerNodeID
		e.term = peerTerm
		e.isLeader = false
		e.notify()
	}
}

func (e *memoryElector) CurrentTermAndVote() (int64, string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.term, e.votedFor, e.isLeader
}


func loadTerm(file string) int64 {
	data, err := os.ReadFile(file)
	if err != nil {
		return 0
	}
	val, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return 0
	}
	return val
}

func (e *memoryElector) saveTerm() {
	tmpFile := e.termFile + ".tmp"
	if err := os.WriteFile(tmpFile, []byte(strconv.FormatInt(e.term, 10)), 0644); err == nil {
		os.Rename(tmpFile, e.termFile)
	}
}

func (e *memoryElector) Campaign(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.term++
	e.votedFor = e.nodeID
	e.votesGranted = 1
	e.saveTerm()

	// Special case for single node clusters
	if e.nodesCount <= 1 {
		e.isLeader = true
		e.leaderID = e.nodeID
		e.notify()
	}

	// Note: in the real system, the Elector doesn't broadcast votes itself,
	// instead `ReceivePeerState` is triggered by Heartbeater, and Heartbeater
	// reads the Elector's state (via CurrentTermAndVote) to broadcast.

	return nil
}

func (e *memoryElector) Resign(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isLeader {
		e.isLeader = false
		e.leaderID = ""
		e.notify()
	}
	return nil
}

func (e *memoryElector) IsLeader() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.isLeader
}

func (e *memoryElector) Status() LeaderStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return LeaderStatus{
		LeaderID:  e.leaderID,
		IsLeader:  e.isLeader,
		Term:      e.term,
		UpdatedAt: time.Now(),
	}
}

func (e *memoryElector) Watch() <-chan LeaderStatus {
	return e.watcher.Subscribe()
}

func (e *memoryElector) Close() error {
	e.watcher.Close()
	return nil
}

func (e *memoryElector) OnLeaderChange(callback func(info LeaderInfo)) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.legacyCbs = append(e.legacyCbs, callback)
}

func (e *memoryElector) notify() {
	status := LeaderStatus{
		LeaderID:  e.leaderID,
		IsLeader:  e.isLeader,
		Term:      e.term,
		UpdatedAt: time.Now(),
	}
	e.watcher.Notify(status)

	info := LeaderInfo{
		NodeID: e.leaderID,
		Term:   uint64(e.term),
	}
	for _, cb := range e.legacyCbs {
		cb(info)
	}
}

