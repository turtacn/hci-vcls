package election

import "sync"

type Watcher struct {
	mu          sync.Mutex
	subscribers []chan LeaderStatus
	closed      bool
}

func NewWatcher() *Watcher {
	return &Watcher{
		subscribers: make([]chan LeaderStatus, 0),
	}
}

func (w *Watcher) Subscribe() <-chan LeaderStatus {
	w.mu.Lock()
	defer w.mu.Unlock()

	ch := make(chan LeaderStatus, 10)
	w.subscribers = append(w.subscribers, ch)
	return ch
}

func (w *Watcher) Notify(status LeaderStatus) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return
	}

	for _, ch := range w.subscribers {
		select {
		case ch <- status:
		default:
		}
	}
}

func (w *Watcher) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return
	}
	w.closed = true

	for _, ch := range w.subscribers {
		close(ch)
	}
	w.subscribers = nil
}

