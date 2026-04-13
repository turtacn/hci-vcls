package election

import (
	"context"
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
	termStore    TermStore
	cancel       context.CancelFunc

	saveCh chan termState
}

var _ Elector = &memoryElector{}

func NewMemoryElector(nodeID string, store TermStore) *memoryElector {
	term := int64(0)
	votedFor := ""
	if store != nil {
		term, votedFor, _ = store.Load()
	}

	e := &memoryElector{
		nodeID:     nodeID,
		term:       term,
		votedFor:   votedFor,
		watcher:    NewWatcher(),
		legacyCbs:  make([]func(LeaderInfo), 0),
		termStore:  store,
		nodesCount: 3, // default to 3 nodes for testing if not set dynamically
		saveCh:     make(chan termState, 100),
	}

	if store != nil {
		go e.saveLoop()
	}

	return e
}

func (e *memoryElector) saveLoop() {
	for state := range e.saveCh {
		_ = e.termStore.Save(state.Term, state.VotedFor)
	}
}

func (e *memoryElector) SetNodesCount(count int) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nodesCount = count
}

func (e *memoryElector) asyncSaveTerm(term int64, votedFor string) {
	if e.termStore != nil {
		select {
		case e.saveCh <- termState{Term: term, VotedFor: votedFor}:
		default:
		}
	}
}

// Receives peer state, optionally incorporating dynamic VoteWeight for witness
func (e *memoryElector) ReceivePeerStateWithWeight(peerNodeID string, peerTerm int64, peerVoteFor string, isLeader bool, voteWeight int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// If peer term is higher, we step down to follower
	if peerTerm > e.term {
		e.term = peerTerm
		e.isLeader = false
		e.leaderID = ""
		e.votedFor = ""
		e.asyncSaveTerm(e.term, e.votedFor)
		e.notify()
	}

	// Simple voting logic
	// If a peer is campaigning, it votes for itself (peerVoteFor == peerNodeID)
	if peerVoteFor == peerNodeID && peerTerm >= e.term && (e.votedFor == "" || e.votedFor == peerNodeID) {
		// Vote for the peer
		e.votedFor = peerNodeID
		e.term = peerTerm
		e.asyncSaveTerm(e.term, e.votedFor)
	} else if peerVoteFor == e.nodeID && peerTerm == e.term {
		// They voted for us
		e.votesGranted += voteWeight
		if !e.isLeader {
			requiredVotes := (e.nodesCount / 2) + 1
			if e.votesGranted >= requiredVotes {
				e.isLeader = true
				e.leaderID = e.nodeID
				e.notify()
			}
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

func (e *memoryElector) ReceivePeerState(peerNodeID string, peerTerm int64, peerVoteFor string, isLeader bool) {
	e.ReceivePeerStateWithWeight(peerNodeID, peerTerm, peerVoteFor, isLeader, 1)
}

func (e *memoryElector) CurrentTermAndVote() (int64, string, bool) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.term, e.votedFor, e.isLeader
}

func (e *memoryElector) Campaign(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.term++
	e.votedFor = e.nodeID
	e.votesGranted = 1 // Vote weight of 1 for self
	e.asyncSaveTerm(e.term, e.votedFor)

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
	if e.termStore != nil {
		close(e.saveCh)
	}
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
