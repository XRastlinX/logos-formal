// Status: PROPOSED
// authority_effect: NONE
// Package outbox defines the transmitting state machine queue adapters.
package outbox

import (
	"fmt"
	"sync"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	cyerrors "github.com/XRastlinX/logos-formal/pkg/cyexchange/errors"
)

type State int

const (
	StateUnknown      State = iota
	StateQueued             // 1
	StateTransmitting       // 2
	StateDelivered          // 3
	StateFailed             // 4
)

// Adapter simulates the outbox durable queue.
type Adapter struct {
	mu    sync.RWMutex
	store map[string]*Record
}

type Record struct {
	Envelope *envelope.CyExchangeEnvelope
	State    State
	QueuedAt time.Time
}

func NewAdapter() *Adapter {
	return &Adapter{
		store: make(map[string]*Record),
	}
}

// Enqueue adds an envelope to the outbox for transmission.
func (a *Adapter) Enqueue(env *envelope.CyExchangeEnvelope) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if _, exists := a.store[env.EnvelopeID]; exists {
		return fmt.Errorf("%w: envelope already queued", cyerrors.ErrReplay)
	}

	a.store[env.EnvelopeID] = &Record{
		Envelope: env,
		State:    StateQueued,
		QueuedAt: time.Now(),
	}
	return nil
}

// MarkTransmitting shifts the envelope state.
func (a *Adapter) MarkTransmitting(envelopeID string) error {
	return a.transition(envelopeID, StateTransmitting)
}

// MarkDelivered signals a cryptographic receipt was successfully parsed.
func (a *Adapter) MarkDelivered(envelopeID string) error {
	return a.transition(envelopeID, StateDelivered)
}

// MarkFailed signals transmission exhaustion or explicit rejection.
func (a *Adapter) MarkFailed(envelopeID string) error {
	return a.transition(envelopeID, StateFailed)
}

func (a *Adapter) transition(envelopeID string, newState State) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	rec, exists := a.store[envelopeID]
	if !exists {
		return fmt.Errorf("record not found in outbox")
	}

	if rec.State == StateDelivered || rec.State == StateFailed {
		return fmt.Errorf("%w: terminal outbox state", cyerrors.ErrValidation)
	}

	if newState <= rec.State {
		return fmt.Errorf("%w: illegal backward transition", cyerrors.ErrValidation)
	}

	rec.State = newState
	return nil
}
