package election

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestTermStore_SaveLoad_Roundtrip(t *testing.T) {
	dir, err := os.MkdirTemp("", "termstore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "term.gob")
	store, err := NewGobTermStore(path)
	if err != nil {
		t.Fatal(err)
	}

	err = store.Save(42, "node-1")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	term, votedFor, err := store.Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if term != 42 || votedFor != "node-1" {
		t.Errorf("Expected 42, node-1; got %d, %s", term, votedFor)
	}
}

func TestTermStore_CorruptFile_ReturnsError(t *testing.T) {
	dir, err := os.MkdirTemp("", "termstore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "term.gob")
	store, err := NewGobTermStore(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte("garbage"), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, err = store.Load()
	if err == nil {
		t.Error("Expected error on corrupt file, got nil")
	}
}

func TestTermStore_ConcurrentSave_NoRace(t *testing.T) {
	dir, err := os.MkdirTemp("", "termstore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "term.gob")
	store, err := NewGobTermStore(path)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(term int64) {
			defer wg.Done()
			_ = store.Save(term, "node-x")
		}(int64(i))
	}
	wg.Wait()
}

func TestTermStore_AtomicRename_NoPartialWrite(t *testing.T) {
	dir, err := os.MkdirTemp("", "termstore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "term.gob")
	store, err := NewGobTermStore(path)
	if err != nil {
		t.Fatal(err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte("garbage"), 0644); err != nil {
		t.Fatal(err)
	}

	term, _, err := store.Load()
	if err != nil {
		t.Fatalf("Load should ignore tmp file, got error: %v", err)
	}
	if term != 0 {
		t.Errorf("Expected 0 term, got %d", term)
	}
}

func TestTermStore_EmptyFile_ReturnsZeroAndNil(t *testing.T) {
	dir, err := os.MkdirTemp("", "termstore-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	path := filepath.Join(dir, "term.gob")
	store, err := NewGobTermStore(path)
	if err != nil {
		t.Fatal(err)
	}

	term, votedFor, err := store.Load()
	if err != nil {
		t.Fatalf("Load on non-existent file failed: %v", err)
	}
	if term != 0 || votedFor != "" {
		t.Errorf("Expected 0 and empty string, got %d, %s", term, votedFor)
	}
}
