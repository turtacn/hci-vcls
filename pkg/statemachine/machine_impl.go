package statemachine

import (
	"context"
	"time"

	"github.com/turtacn/hci-vcls/pkg/fdm"
)

type machineImpl struct {
	config     StateMachineConfig
	prober     *HealthProber
	current    fdm.DegradationLevel
	lastTrans  Transition
	history    []Transition
	callbacks  []func(fdm.DegradationLevel, fdm.DegradationLevel, string)
	ctx        context.Context
	cancel     context.CancelFunc
	lastEvalAt time.Time
}

func NewMachine(config StateMachineConfig, prober *HealthProber) Machine {
	ctx, cancel := context.WithCancel(context.Background())
	return &machineImpl{
		config:    config,
		prober:    prober,
		current:   fdm.DegradationNone,
		history:   make([]Transition, 0),
		callbacks: make([]func(fdm.DegradationLevel, fdm.DegradationLevel, string), 0),
		ctx:       ctx,
		cancel:    cancel,
	}
}

func (m *machineImpl) Start(ctx context.Context) error {
	go m.evalLoop()
	return nil
}

func (m *machineImpl) Stop() error {
	m.cancel()
	return nil
}

func (m *machineImpl) CurrentLevel() fdm.DegradationLevel {
	return m.current
}

func (m *machineImpl) LastTransition() Transition {
	return m.lastTrans
}

func (m *machineImpl) TransitionHistory() []Transition {
	return m.history
}

func (m *machineImpl) OnTransition(callback func(from, to fdm.DegradationLevel, reason string)) {
	m.callbacks = append(m.callbacks, callback)
}

func (m *machineImpl) ForceEvaluate(ctx context.Context) (EvaluationResult, error) {
	if time.Since(m.lastEvalAt) < time.Duration(m.config.CooldownMs)*time.Millisecond {
		return EvaluationResult{}, ErrCooldownActive
	}

	input := m.prober.Sample(ctx)
	result := Evaluate(input)
	m.applyTransition(result.Level, result.Reason)
	m.lastEvalAt = time.Now()

	return result, nil
}

func (m *machineImpl) evalLoop() {
	ticker := time.NewTicker(time.Duration(m.config.EvaluationIntervalMs) * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			if time.Since(m.lastEvalAt) >= time.Duration(m.config.CooldownMs)*time.Millisecond {
				input := m.prober.Sample(m.ctx)
				result := Evaluate(input)
				m.applyTransition(result.Level, result.Reason)
				m.lastEvalAt = time.Now()
			}
		}
	}
}

func (m *machineImpl) applyTransition(to fdm.DegradationLevel, reason string) {
	if m.current != to {
		if IsValidTransition(m.current, to) {
			trans := Transition{From: m.current, To: to, Timestamp: time.Now(), Reason: reason}
			m.lastTrans = trans
			m.history = append(m.history, trans)
			m.current = to
			for _, cb := range m.callbacks {
				cb(trans.From, trans.To, trans.Reason)
			}
		}
	}
}

//Personal.AI order the ending