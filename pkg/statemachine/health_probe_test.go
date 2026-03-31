package statemachine

import (
	"context"
	"testing"

	"github.com/turtacn/hci-vcls/pkg/cfs"
	"github.com/turtacn/hci-vcls/pkg/fdm"
	"github.com/turtacn/hci-vcls/pkg/mysql"
	"github.com/turtacn/hci-vcls/pkg/zk"
)

type mockCFSAdapter struct {
	state cfs.HealthState
}

func (m *mockCFSAdapter) Health() cfs.CFSStatus               { return cfs.CFSStatus{State: m.state} }
func (m *mockCFSAdapter) IsWritable() cfs.CFSStatus           { return cfs.CFSStatus{State: cfs.CFSStateHealthy} }
func (m *mockCFSAdapter) ReadVMConfig(vmid string) ([]byte, error) { return nil, nil }
func (m *mockCFSAdapter) ListVMIDs() ([]string, error)        { return nil, nil }
func (m *mockCFSAdapter) Close() error                        { return nil }

type mockMySQLAdapter struct {
	state mysql.HealthState
}

func (m *mockMySQLAdapter) Health() mysql.MySQLStatus                      { return mysql.MySQLStatus{State: m.state} }
func (m *mockMySQLAdapter) ClaimBoot(claim mysql.BootClaim) error          { return nil }
func (m *mockMySQLAdapter) ConfirmBoot(vmid, token string) error           { return nil }
func (m *mockMySQLAdapter) ReleaseBoot(vmid, token string) error           { return nil }
func (m *mockMySQLAdapter) GetVMState(vmid string) (*mysql.HAVMState, error) { return nil, nil }
func (m *mockMySQLAdapter) UpsertVMState(state mysql.HAVMState) error      { return nil }
func (m *mockMySQLAdapter) Close() error                                   { return nil }

type mockFDMAgent struct {
	level fdm.DegradationLevel
}

func (m *mockFDMAgent) Start(ctx context.Context) error                         { return nil }
func (m *mockFDMAgent) Stop() error                                             { return nil }
func (m *mockFDMAgent) NodeStates() map[string]fdm.NodeState                    { return nil }
func (m *mockFDMAgent) LocalDegradationLevel() fdm.DegradationLevel             { return m.level }
func (m *mockFDMAgent) IsLeader() bool                                          { return false }
func (m *mockFDMAgent) LeaderNodeID() string                                    { return "" }
func (m *mockFDMAgent) OnNodeFailure(callback func(nodeID string))              {}
func (m *mockFDMAgent) OnDegradationChanged(callback func(level fdm.DegradationLevel)) {}
func (m *mockFDMAgent) ClusterView() fdm.ClusterView                            { return fdm.ClusterView{} }

func TestHealthProber_Sample(t *testing.T) {
	hp := NewHealthProber(
		&mockZKAdapter{state: zk.ZKStateUnavailable},
		&mockCFSAdapter{state: cfs.CFSStateHealthy},
		&mockMySQLAdapter{state: mysql.MySQLStateHealthy},
		&mockFDMAgent{level: fdm.DegradationNone},
	)

	ctx := context.Background()
	input := hp.Sample(ctx)

	if input.ZKStatus.State != zk.ZKStateUnavailable {
		t.Errorf("Expected ZK Unavailable, got %v", input.ZKStatus.State)
	}
	if input.CFSStatus.State != cfs.CFSStateHealthy {
		t.Errorf("Expected CFS Healthy, got %v", input.CFSStatus.State)
	}
	if input.MySQLStatus.State != mysql.MySQLStateHealthy {
		t.Errorf("Expected MySQL Healthy, got %v", input.MySQLStatus.State)
	}
	if input.FDMLevel != fdm.DegradationNone {
		t.Errorf("Expected FDM DegradationNone, got %v", input.FDMLevel)
	}
}

//Personal.AI order the ending