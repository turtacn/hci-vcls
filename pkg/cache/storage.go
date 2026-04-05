package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type StorageStore interface {
	Get(vmid string) (*VMStorageMeta, error)
	Put(vmid string, meta VMStorageMeta) error
	Delete(vmid string) error
	List() ([]VMStorageMeta, error)
}

type StorageStoreConfig struct {
	BaseDir string
}

type storageStoreImpl struct {
	baseDir string
}

func NewStorageStore(baseDir string) (StorageStore, error) {
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		return nil, err
	}
	return &storageStoreImpl{baseDir: baseDir}, nil
}

func (s *storageStoreImpl) Get(vmid string) (*VMStorageMeta, error) {
	path := filepath.Join(s.baseDir, vmid+"_storage.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	var meta VMStorageMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *storageStoreImpl) Put(vmid string, meta VMStorageMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	path := filepath.Join(s.baseDir, vmid+"_storage.json")
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func (s *storageStoreImpl) Delete(vmid string) error {
	path := filepath.Join(s.baseDir, vmid+"_storage.json")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *storageStoreImpl) List() ([]VMStorageMeta, error) {
	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}

	var metas []VMStorageMeta
	for _, f := range files {
		if !f.IsDir() {
			name := f.Name()
			if len(name) > 13 && name[len(name)-13:] == "_storage.json" {
				vmid := name[:len(name)-13]
				meta, err := s.Get(vmid)
				if err == nil && meta != nil {
					metas = append(metas, *meta)
				}
			}
		}
	}
	return metas, nil
}

// Personal.AI order the ending
