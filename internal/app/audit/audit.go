package audit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type AuditLogger interface {
	LogHADecision(ctx context.Context, d HADecisionRecord) error
	Close() error
}

type HADecisionRecord struct {
	Timestamp   time.Time `json:"timestamp"`
	PlanID      string    `json:"plan_id"`
	VMID        string    `json:"vm_id"`
	ClusterID   string    `json:"cluster_id"`
	BootPath    string    `json:"boot_path"`
	TargetHost  string    `json:"target_host"`
	SourceHost  string    `json:"source_host"`
	Reason      string    `json:"reason"`
	Degradation string    `json:"degradation"`
	Outcome     string    `json:"outcome"`
	Error       string    `json:"error,omitempty"`
}

type jsonlAuditLogger struct {
	mu       sync.Mutex
	dir      string
	file     *os.File
	currDate string
}

func NewJSONLAuditLogger(dir string) (AuditLogger, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &jsonlAuditLogger{dir: dir}, nil
}

func (l *jsonlAuditLogger) getFilename(date string) string {
	return filepath.Join(l.dir, "ha-decisions-"+date+".jsonl")
}

func (l *jsonlAuditLogger) rotateIfNeeded() error {
	nowDate := time.Now().Format("2006-01-02")
	if l.file != nil && l.currDate == nowDate {
		return nil
	}

	if l.file != nil {
		_ = l.file.Sync()
		_ = l.file.Close()
	}

	path := l.getFilename(nowDate)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	l.file = f
	l.currDate = nowDate
	return nil
}

func (l *jsonlAuditLogger) LogHADecision(ctx context.Context, d HADecisionRecord) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if d.Timestamp.IsZero() {
		d.Timestamp = time.Now()
	}

	if err := l.rotateIfNeeded(); err != nil {
		return err
	}

	b, err := json.Marshal(d)
	if err != nil {
		return err
	}

	b = append(b, '\n')
	_, err = l.file.Write(b)
	if err != nil {
		return err
	}

	return l.file.Sync()
}

func (l *jsonlAuditLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.file != nil {
		err := l.file.Sync()
		closeErr := l.file.Close()
		l.file = nil
		if err != nil {
			return err
		}
		return closeErr
	}
	return nil
}
