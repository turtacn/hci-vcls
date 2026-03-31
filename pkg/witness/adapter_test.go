package witness

import (
	"context"
	"testing"
)

type mockLogger struct{}

func (m mockLogger) Debug(msg string, args ...any)            {}
func (m mockLogger) Info(msg string, args ...any)             {}
func (m mockLogger) Warn(msg string, args ...any)             {}
func (m mockLogger) Error(msg string, args ...any)            {}
func (m mockLogger) WithFields(fields ...any) interface{} { return m }
func (m mockLogger) WithError(err error) interface{}      { return m }

func TestPool_ConfirmFailure(t *testing.T) {
	ctx := context.Background()
	log := mockLogger{}

	config := WitnessConfig{Endpoints: []string{"http://w1:8080"}}
	pool, _ := NewPool(config, nil)

	req := ConfirmationRequest{NodeID: "node-1"}
	confirmed := pool.ConfirmFailure(ctx, req)

	if !confirmed {
		t.Errorf("Expected failure to be confirmed with 1 endpoint")
	}

	configEmpty := WitnessConfig{Endpoints: []string{}}
	poolEmpty, _ := NewPool(configEmpty, log)

	confirmedEmpty := poolEmpty.ConfirmFailure(ctx, req)
	if confirmedEmpty {
		t.Errorf("Expected failure not to be confirmed with 0 endpoints")
	}
}

func TestPool_Quorum(t *testing.T) {
	ctx := context.Background()

	config := WitnessConfig{Endpoints: []string{"http://w1:8080", "http://w2:8080"}}
	pool, _ := NewPool(config, nil)

	if !pool.Quorum(ctx) {
		t.Errorf("Expected quorum to be true")
	}

	configEmpty := WitnessConfig{Endpoints: []string{}}
	poolEmpty, _ := NewPool(configEmpty, nil)

	if poolEmpty.Quorum(ctx) {
		t.Errorf("Expected quorum to be false")
	}
}

func TestPool_Statuses(t *testing.T) {
	ctx := context.Background()
	config := WitnessConfig{Endpoints: []string{"http://w1:8080", "http://w2:8080"}}
	pool, _ := NewPool(config, nil)

	statuses := pool.Statuses(ctx)

	if len(statuses) != 2 {
		t.Errorf("Expected 2 statuses, got %d", len(statuses))
	}

	for _, endpoint := range config.Endpoints {
		if status, ok := statuses[endpoint]; !ok || status != StatusHealthy {
			t.Errorf("Expected healthy status for endpoint %s", endpoint)
		}
	}
}

func TestAdapter_Health(t *testing.T) {
	ctx := context.Background()

	config := WitnessConfig{Endpoints: []string{"http://w1:8080"}}
	adapter, _ := NewAdapter(config, nil)

	status := adapter.Health(ctx)
	if status != StatusHealthy {
		t.Errorf("Expected healthy status, got %v", status)
	}

	configEmpty := WitnessConfig{Endpoints: []string{}}
	adapterEmpty, _ := NewAdapter(configEmpty, nil)

	statusEmpty := adapterEmpty.Health(ctx)
	if statusEmpty != StatusUnknown {
		t.Errorf("Expected unknown status, got %v", statusEmpty)
	}
}

func TestAdapter_ConfirmFailure(t *testing.T) {
	ctx := context.Background()
	config := WitnessConfig{Endpoints: []string{"http://w1:8080"}}
	adapter, _ := NewAdapter(config, nil)

	req := ConfirmationRequest{NodeID: "node-1"}
	resp := adapter.ConfirmFailure(ctx, req)

	if !resp.Confirmed {
		t.Errorf("Expected confirmation to be true")
	}
	if resp.Error != nil {
		t.Errorf("Expected no error, got %v", resp.Error)
	}
}

//Personal.AI order the ending