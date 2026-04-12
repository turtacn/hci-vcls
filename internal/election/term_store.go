package election

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"sync"
)

type TermStore interface {
	Load() (term int64, votedFor string, err error)
	Save(term int64, votedFor string) error
	Close() error
}

type termState struct {
	Term     int64  `json:"term"`
	VotedFor string `json:"voted_for"`
	Version  uint32 `json:"version"`
}

type gobTermStore struct {
	mu   sync.Mutex
	path string
}

func NewGobTermStore(path string) (TermStore, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &gobTermStore{path: path}, nil
}

func (s *gobTermStore) Load() (int64, string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, "", nil
		}
		return 0, "", err
	}
	defer f.Close()

	var state termState
	decoder := gob.NewDecoder(f)
	if err := decoder.Decode(&state); err != nil {
		return 0, "", err
	}

	return state.Term, state.VotedFor, nil
}

func (s *gobTermStore) Save(term int64, votedFor string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	tmpPath := s.path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	defer func() {
		f.Close()
		os.Remove(tmpPath)
	}()

	state := termState{
		Term:     term,
		VotedFor: votedFor,
		Version:  1,
	}

	encoder := gob.NewEncoder(f)
	if err := encoder.Encode(&state); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		return err
	}

	dirPath := filepath.Dir(s.path)
	d, err := os.Open(dirPath)
	if err != nil {
		return err
	}
	defer d.Close()

	if err := d.Sync(); err != nil {
		return err
	}

	return nil
}

func (s *gobTermStore) Close() error {
	return nil
}
