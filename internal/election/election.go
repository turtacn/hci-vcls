package election

import (
	"context"
	"sync"
	"time"
)

var (
	globalLeaderLock sync.Mutex
	currentLeader    string
)

type memoryElector struct {
	mu        sync.RWMutex
	nodeID    string
	isLeader  bool
	term      int64
	watcher   *Watcher
	legacyCbs []func(LeaderInfo)
}

var _ Elector = &memoryElector{}

func NewMemoryElector(nodeID string) *memoryElector {
	return &memoryElector{
		nodeID:    nodeID,
		watcher:   NewWatcher(),
		legacyCbs: make([]func(LeaderInfo), 0),
	}
}

func (e *memoryElector) Campaign(ctx context.Context) error {
	globalLeaderLock.Lock()
	defer globalLeaderLock.Unlock()

	e.mu.Lock()
	defer e.mu.Unlock()

	if currentLeader == "" || currentLeader == e.nodeID {
		currentLeader = e.nodeID
		e.isLeader = true
		e.term++
		e.notify()
	}
	return nil
}

func (e *memoryElector) Resign(ctx context.Context) error {
	globalLeaderLock.Lock()
	defer globalLeaderLock.Unlock()

	e.mu.Lock()
	defer e.mu.Unlock()

	if currentLeader == e.nodeID {
		currentLeader = ""
		e.isLeader = false
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
		LeaderID:  currentLeader,
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
		LeaderID:  currentLeader,
		IsLeader:  e.isLeader,
		Term:      e.term,
		UpdatedAt: time.Now(),
	}
	e.watcher.Notify(status)

	info := LeaderInfo{
		NodeID: currentLeader,
		Term:   uint64(e.term),
	}
	for _, cb := range e.legacyCbs {
		cb(info)
	}
}

// Personal.AI order the ending
