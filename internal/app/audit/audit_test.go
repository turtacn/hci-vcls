package audit

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestJSONLAuditLogger_LogAndRead(t *testing.T) {
	dir := t.TempDir()

	logger, err := NewJSONLAuditLogger(dir)
	if err != nil {
		t.Fatalf("unexpected error creating logger: %v", err)
	}

	record := HADecisionRecord{
		PlanID:      "plan-1",
		VMID:        "vm-1",
		ClusterID:   "cluster-1",
		BootPath:    "normal",
		TargetHost:  "node-2",
		SourceHost:  "node-1",
		Reason:      "node-1 failed",
		Degradation: "None",
		Outcome:     "success",
	}

	err = logger.LogHADecision(context.Background(), record)
	if err != nil {
		t.Fatalf("unexpected error logging decision: %v", err)
	}

	err = logger.Close()
	if err != nil {
		t.Fatalf("unexpected error closing logger: %v", err)
	}

	date := time.Now().Format("2006-01-02")
	path := filepath.Join(dir, "ha-decisions-"+date+".jsonl")

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("unexpected error opening log file: %v", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		t.Fatalf("expected to read a line from log file")
	}

	var readRecord HADecisionRecord
	err = json.Unmarshal(scanner.Bytes(), &readRecord)
	if err != nil {
		t.Fatalf("unexpected error unmarshaling log line: %v", err)
	}

	if readRecord.PlanID != record.PlanID || readRecord.VMID != record.VMID || readRecord.Outcome != record.Outcome {
		t.Errorf("read record does not match written record: got %+v", readRecord)
	}
}

func TestJSONLAuditLogger_ConcurrentWrite(t *testing.T) {
	dir := t.TempDir()

	logger, err := NewJSONLAuditLogger(dir)
	if err != nil {
		t.Fatalf("unexpected error creating logger: %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			record := HADecisionRecord{
				PlanID:  "plan-1",
				VMID:    "vm-1",
				Outcome: "success",
			}
			_ = logger.LogHADecision(context.Background(), record)
		}(i)
	}
	wg.Wait()

	_ = logger.Close()

	date := time.Now().Format("2006-01-02")
	path := filepath.Join(dir, "ha-decisions-"+date+".jsonl")

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("unexpected error opening log file: %v", err)
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
	}

	if count != 100 {
		t.Errorf("expected 100 lines, got %d", count)
	}
}
