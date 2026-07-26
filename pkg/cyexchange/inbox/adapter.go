// Status: PROPOSED
// authority_effect: NONE
// Package inbox defines the state machine queue adapters.
package inbox

import (
	"fmt"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
)

// State represents the CyExchange state machine state of an envelope.
type State string

const (
	StateReceived        State = "RECEIVED"
	StateEvaluating      State = "EVALUATING"
	StateObserveOnly     State = "OBSERVE_ONLY"
	StateEffectRequested State = "EFFECT_REQUESTED"
	StateApplied         State = "APPLIED"
	StateRejected        State = "REJECTED"
	StateOutcomeUnknown  State = "OUTCOME_UNKNOWN"
)

// Adapter simulates the local durable storage for incoming envelopes.
type Adapter struct {
	store map[string]*Record
}

// Record binds an envelope to its state machine state.
type Record struct {
	Envelope *envelope.CyExchangeEnvelope
	State    State
}

// NewAdapter creates a new in-memory inbox adapter.
func NewAdapter() *Adapter {
	return &Adapter{
		store: make(map[string]*Record),
	}
}

// Ingest accepts a new envelope and persists it, returning a delivery receipt stub.
func (a *Adapter) Ingest(env *envelope.CyExchangeEnvelope) error {
	if _, exists := a.store[env.EnvelopeID]; exists {
		return fmt.Errorf("idempotency violation: envelope %s already exists", env.EnvelopeID)
	}

	a.store[env.EnvelopeID] = &Record{
		Envelope: env,
		State:    StateReceived,
	}
	return nil
}

// Transition safely moves an envelope between states.
func (a *Adapter) Transition(envelopeID string, newState State) error {
	record, exists := a.store[envelopeID]
	if !exists {
		return fmt.Errorf("record not found")
	}

	// Strictly monotonic checks can be placed here in ACTIVE_CANON
	if record.State == StateApplied {
		return fmt.Errorf("cannot transition from terminal state APPLIED")
	}

	record.State = newState
	return nil
}
