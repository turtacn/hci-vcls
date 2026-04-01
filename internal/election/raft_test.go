package election

import (
	"context"
	"errors"
	"testing"

	"github.com/turtacn/hci-vcls/pkg/zk"
)

type mockZKAdapter struct{}

func (m mockZKAdapter) Health() zk.ZKStatus     { return zk.ZKStatus{State: zk.ZKStateHealthy} }
func (m mockZKAdapter) IsReadOnly() zk.ZKStatus { return zk.ZKStatus{State: zk.ZKStateHealthy} }
func (m mockZKAdapter) Ping() zk.ZKStatus       { return zk.ZKStatus{State: zk.ZKStateHealthy} }
func (m mockZKAdapter) Close() error            { return nil }

func TestElector_CampaignResign(t *testing.T) {
	config := ElectionConfig{NodeID: "node-1"}
	elector := NewElector(config, mockZKAdapter{})

	ctx := context.Background()

	// Initial state
	if elector.IsLeader() {
		t.Errorf("Expected not to be leader initially")
	}

	leaderID, err := elector.LeaderID()
	if err != ErrNoLeader {
		t.Errorf("Expected ErrNoLeader, got %v", err)
	}

	var notifiedLeaderID string
	elector.OnLeaderChange(func(info LeaderInfo) {
		notifiedLeaderID = info.NodeID
	})

	// Campaign
	err = elector.Campaign(ctx)
	if err != nil {
		t.Errorf("Expected campaign to succeed, got %v", err)
	}

	if !elector.IsLeader() {
		t.Errorf("Expected to be leader after campaign")
	}

	leaderID, err = elector.LeaderID()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if leaderID != "node-1" {
		t.Errorf("Expected leaderID 'node-1', got '%s'", leaderID)
	}

	if notifiedLeaderID != "node-1" {
		t.Errorf("Expected callback with 'node-1', got '%s'", notifiedLeaderID)
	}

	// Resign
	err = elector.Resign(ctx)
	if err != nil {
		t.Errorf("Expected resign to succeed, got %v", err)
	}

	if elector.IsLeader() {
		t.Errorf("Expected not to be leader after resign")
	}

	leaderID, err = elector.LeaderID()
	if err != ErrNoLeader {
		t.Errorf("Expected ErrNoLeader, got %v", err)
	}
}

func TestElector_Close(t *testing.T) {
	config := ElectionConfig{NodeID: "node-1"}
	elector := NewElector(config, mockZKAdapter{})

	ctx := context.Background()

	err := elector.Campaign(ctx)
	if err != nil {
		t.Errorf("Expected campaign to succeed, got %v", err)
	}

	if !elector.IsLeader() {
		t.Errorf("Expected to be leader after campaign")
	}

	err = elector.Close()
	if err != nil {
		t.Errorf("Expected close to succeed, got %v", err)
	}

	if elector.IsLeader() {
		t.Errorf("Expected not to be leader after close")
	}
}

func TestElectionError(t *testing.T) {
	err := ErrNoLeader
	if err.Error() == "" {
		t.Errorf("Expected error message to not be empty")
	}

	wrappedErr := &ElectionError{Code: "ERR_CUSTOM", Message: "custom error", Err: errors.New("inner error")}
	if err := wrappedErr.Unwrap(); err == nil {
		t.Errorf("Expected to unwrap an inner error")
	}
}

//Personal.AI order the ending
