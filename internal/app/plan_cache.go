package app

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/turtacn/hci-vcls/pkg/ha"
)

type PlanCache interface {
	Put(plan *ha.Plan) error
	Get(planID string) (*ha.Plan, error)
	Delete(planID string) error
	List() ([]*ha.Plan, error)
}

type fsPlanCache struct {
	mu  sync.Mutex
	dir string
}

func NewFSPlanCache(dir string) (PlanCache, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &fsPlanCache{dir: dir}, nil
}

func (c *fsPlanCache) getPath(planID string) string {
	return filepath.Join(c.dir, planID+".json")
}

func (c *fsPlanCache) Put(plan *ha.Plan) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.getPath(plan.ID)
	tmpPath := path + ".tmp"

	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
		os.Remove(tmpPath)
	}()

	encoder := json.NewEncoder(f)
	if err := encoder.Encode(plan); err != nil {
		return err
	}

	if err := f.Sync(); err != nil {
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return err
	}

	d, err := os.Open(c.dir)
	if err != nil {
		return err
	}
	defer d.Close()

	if err := d.Sync(); err != nil {
		return err
	}

	return nil
}

func (c *fsPlanCache) Get(planID string) (*ha.Plan, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.getPath(planID)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var plan ha.Plan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, err
	}

	return &plan, nil
}

func (c *fsPlanCache) Delete(planID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	path := c.getPath(planID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (c *fsPlanCache) List() ([]*ha.Plan, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var plans []*ha.Plan

	err := filepath.WalkDir(c.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil // skip unreadable files
		}

		var plan ha.Plan
		if err := json.Unmarshal(data, &plan); err == nil {
			plans = append(plans, &plan)
		}
		return nil
	})

	return plans, err
}
