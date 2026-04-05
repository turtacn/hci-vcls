package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type NetworkStore interface {
	Get(vmid string) (*VMNetworkMeta, error)
	Put(vmid string, meta VMNetworkMeta) error
	Delete(vmid string) error
	List() ([]VMNetworkMeta, error)
}

type NetworkStoreConfig struct {
	BaseDir string
}

type networkStoreImpl struct {
	baseDir string
}

func NewNetworkStore(baseDir string) (NetworkStore, error) {
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		return nil, err
	}
	return &networkStoreImpl{baseDir: baseDir}, nil
}

func (s *networkStoreImpl) Get(vmid string) (*VMNetworkMeta, error) {
	path := filepath.Join(s.baseDir, vmid+"_net.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrCacheMiss
		}
		return nil, err
	}

	var meta VMNetworkMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *networkStoreImpl) Put(vmid string, meta VMNetworkMeta) error {
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	path := filepath.Join(s.baseDir, vmid+"_net.json")
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func (s *networkStoreImpl) Delete(vmid string) error {
	path := filepath.Join(s.baseDir, vmid+"_net.json")
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *networkStoreImpl) List() ([]VMNetworkMeta, error) {
	files, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}

	var metas []VMNetworkMeta
	for _, f := range files {
		if !f.IsDir() {
			name := f.Name()
			if len(name) > 9 && name[len(name)-9:] == "_net.json" {
				vmid := name[:len(name)-9]
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
