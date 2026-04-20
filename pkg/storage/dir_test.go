package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func TestDirStorage(t *testing.T) {
	tempDir := t.TempDir()

	s := NewDirStorage(tempDir)
	if s.Type() != StorageDir {
		t.Errorf("expected StorageDir, got %v", s.Type())
	}

	ctx := context.Background()

	// 1. Probe successful
	status, err := s.Probe(ctx, "local")
	if err != nil {
		t.Errorf("Probe failed: %v", err)
	}
	if !status.Available {
		t.Errorf("Expected storage to be available")
	}

	// 2. IsAccessible
	accessible, err := s.IsAccessible(ctx, "local", "node-1")
	if err != nil {
		t.Errorf("IsAccessible failed: %v", err)
	}
	if !accessible {
		t.Errorf("Expected storage to be accessible")
	}

	// 3. Mount/Unmount (no-ops)
	if err := s.Mount(ctx, "local", "node-1"); err != nil {
		t.Errorf("Mount failed: %v", err)
	}
	if err := s.Unmount(ctx, "local", "node-1"); err != nil {
		t.Errorf("Unmount failed: %v", err)
	}

	// 4. Probe unavailable
	nonExistentPath := filepath.Join(tempDir, "does-not-exist")
	s2 := NewDirStorage(nonExistentPath)

	status2, err := s2.Probe(ctx, "local")
	if err != nil {
		t.Errorf("Probe should not fail for non-existent dir, got %v", err)
	}
	if status2.Available {
		t.Errorf("Expected storage to be unavailable")
	}
}

func TestRegistry(t *testing.T) {
	reg := NewRegistry()
	tempDir := t.TempDir()

	dirStorage := NewDirStorage(tempDir)
	reg.Register(dirStorage)

	s, err := reg.Get(StorageDir)
	if err != nil {
		t.Errorf("failed to get storage: %v", err)
	}
	if s.Type() != StorageDir {
		t.Errorf("expected StorageDir, got %v", s.Type())
	}

	_, err = reg.Get(StorageRBD)
	if err == nil {
		t.Errorf("expected error getting unregistered storage")
	}

	ctx := context.Background()
	statuses, err := reg.ProbeAll(ctx, "local")
	if err != nil {
		t.Errorf("ProbeAll failed: %v", err)
	}
	if len(statuses) != 1 {
		t.Errorf("expected 1 status, got %d", len(statuses))
	}
}
