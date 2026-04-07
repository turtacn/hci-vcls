package zk

import (
	"context"
	"sync"
)

type MemoryClient struct {
	mu       sync.RWMutex
	nodes    map[string]*ZNode
	state    SessionState
	watchers map[string][]chan WatchEvent
}

var _ Client = &MemoryClient{}

func NewMemoryClient() *MemoryClient {
	return &MemoryClient{
		nodes:    make(map[string]*ZNode),
		state:    Disconnected,
		watchers: make(map[string][]chan WatchEvent),
	}
}

func (c *MemoryClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = Connected
	return nil
}

func (c *MemoryClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = Disconnected
	for _, chs := range c.watchers {
		for _, ch := range chs {
			close(ch)
		}
	}
	c.watchers = make(map[string][]chan WatchEvent)
	return nil
}

func (c *MemoryClient) notify(path string, typ EventType) {
	if chs, ok := c.watchers[path]; ok {
		for _, ch := range chs {
			select {
			case ch <- WatchEvent{Path: path, Type: typ}:
			default:
			}
		}
	}
}

func (c *MemoryClient) Create(path, data string, ephemeral bool) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != Connected {
		return ErrNotConnected
	}

	if _, exists := c.nodes[path]; exists {
		return ErrNodeExists
	}

	c.nodes[path] = &ZNode{
		Path:    path,
		Data:    data,
		Version: 0,
	}

	c.notify(path, EventNodeCreated)
	return nil
}

func (c *MemoryClient) Get(path string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.state != Connected {
		return "", ErrNotConnected
	}

	node, exists := c.nodes[path]
	if !exists {
		return "", ErrNodeNotFound
	}
	return node.Data, nil
}

func (c *MemoryClient) Set(path, data string, version int32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != Connected {
		return ErrNotConnected
	}

	node, exists := c.nodes[path]
	if !exists {
		return ErrNodeNotFound
	}

	if node.Version != version {
		return ErrVersionMismatch
	}

	node.Data = data
	node.Version++
	c.notify(path, EventNodeDataChanged)
	return nil
}

func (c *MemoryClient) Delete(path string, version int32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != Connected {
		return ErrNotConnected
	}

	node, exists := c.nodes[path]
	if !exists {
		return ErrNodeNotFound
	}

	if node.Version != version {
		return ErrVersionMismatch
	}

	delete(c.nodes, path)
	c.notify(path, EventNodeDeleted)
	return nil
}

func (c *MemoryClient) Exists(path string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.state != Connected {
		return false, ErrNotConnected
	}

	_, exists := c.nodes[path]
	return exists, nil
}

func (c *MemoryClient) Watch(path string) (<-chan WatchEvent, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.state != Connected {
		return nil, ErrNotConnected
	}

	ch := make(chan WatchEvent, 10)
	c.watchers[path] = append(c.watchers[path], ch)
	return ch, nil
}

func (c *MemoryClient) SessionState() SessionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

