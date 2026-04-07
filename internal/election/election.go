package election

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
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
	termFile  string
}

var _ Elector = &memoryElector{}

func NewMemoryElector(nodeID string) *memoryElector {
	termDir := "/var/lib/hci-vcls"
	_ = os.MkdirAll(termDir, 0755)
	termFile := filepath.Join(termDir, "hci-vcls-term-"+nodeID)
	term := loadTerm(termFile)

	return &memoryElector{
		nodeID:    nodeID,
		term:      term,
		watcher:   NewWatcher(),
		legacyCbs: make([]func(LeaderInfo), 0),
		termFile:  termFile,
	}
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
	globalLeaderLock.Lock()
	defer globalLeaderLock.Unlock()

	e.mu.Lock()
	defer e.mu.Unlock()

	if currentLeader == "" || currentLeader == e.nodeID {
		currentLeader = e.nodeID
		e.isLeader = true
		e.term++
		e.saveTerm()
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

