// Status: PROPOSED
// authority_effect: NONE
// Package inbox defines the monotonic state machine queue adapters.
package inbox

import (
	"fmt"
	"sync"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	cyerrors "github.com/XRastlinX/logos-formal/pkg/cyexchange/errors"
)

// State represents the CyExchange state machine state of an envelope.
type State int

const (
	StateUnknown         State = iota
	StateReceived              // 1
	StateEvaluating            // 2
	StateObserveOnly           // 3
	StateEffectRequested       // 4
	StateApplied               // 5
	StateRejected              // 6
	StateOutcomeUnknown        // 7
)

// Adapter simulates local durable storage enforcing monotonic transitions.
type Adapter struct {
	mu    sync.RWMutex
	store map[string]*Record
}

// Record binds an envelope to its state machine state.
type Record struct {
	Envelope  *envelope.CyExchangeEnvelope
	State     State
	ExpiresAt time.Time
}

// NewAdapter creates a new in-memory inbox adapter.
func NewAdapter() *Adapter {
	return &Adapter{
		store: make(map[string]*Record),
	}
}

// Ingest accepts a new envelope, checking for deduplication and TTL.
func (a *Adapter) Ingest(env *envelope.CyExchangeEnvelope, ttl time.Duration) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.store[env.EnvelopeID]; exists {
		return fmt.Errorf("%w: envelope %s already exists", cyerrors.ErrReplay, env.EnvelopeID)
	}

	a.store[env.EnvelopeID] = &Record{
		Envelope:  env,
		State:     StateReceived,
		ExpiresAt: time.Now().Add(ttl),
	}
	return nil
}

// Transition safely moves an envelope monotonically between states.
func (a *Adapter) Transition(envelopeID string, newState State) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	record, exists := a.store[envelopeID]
	if !exists {
		return fmt.Errorf("record not found: %s", envelopeID)
	}

	if time.Now().After(record.ExpiresAt) {
		record.State = StateRejected
		return fmt.Errorf("%w: transition attempted on expired envelope", cyerrors.ErrExpiration)
	}

	// Strictly monotonic checks:
	// States 1-5 are sequential. 6 and 7 are terminal.
	if record.State == StateApplied || record.State == StateRejected || record.State == StateOutcomeUnknown {
		return fmt.Errorf("%w: cannot transition from terminal state", cyerrors.ErrValidation)
	}

	if newState <= record.State {
		return fmt.Errorf("%w: illegal backward state transition from %d to %d", cyerrors.ErrValidation, record.State, newState)
	}

	record.State = newState
	return nil
}

// GetState retrieves the current state of an envelope.
func (a *Adapter) GetState(envelopeID string) (State, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	record, exists := a.store[envelopeID]
	if !exists {
		return StateUnknown, fmt.Errorf("record not found")
	}
	return record.State, nil
}
