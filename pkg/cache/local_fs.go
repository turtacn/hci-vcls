package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type localStore struct {
	baseDir string
}

func NewLocalStore(baseDir string) (Store, error) {
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		return nil, err
	}
	return &localStore{baseDir: baseDir}, nil
}

func (s *localStore) Get(vmid string) (*VMComputeMeta, error) {
	path := filepath.Join(s.baseDir, vmid+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	var meta VMComputeMeta
	err = json.Unmarshal(data, &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *localStore) Put(vmid string, meta VMComputeMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	path := filepath.Join(s.baseDir, vmid+".json")
	tmpPath := path + ".tmp"

	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return err
	}

	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return err
	}

	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return err
	}

	return os.Rename(tmpPath, path)
}

func (s *localStore) Delete(vmid string) error {
	path := filepath.Join(s.baseDir, vmid+".json")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *localStore) List() ([]VMComputeMeta, error) {
	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}

	var metas []VMComputeMeta
	for _, f := range files {
		if !f.IsDir() {
			name := f.Name()
			if len(name) > 5 && name[len(name)-5:] == ".json" {
				vmid := name[:len(name)-5]
				meta, err := s.Get(vmid)
				if err == nil && meta != nil {
					metas = append(metas, *meta)
				}
			}
		}
	}
	return metas, nil
}

// ... similar minimal implementations for NetworkStore and StorageStore ...

// Personal.AI order the ending
