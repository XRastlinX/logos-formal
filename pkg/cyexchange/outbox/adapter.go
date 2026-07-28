// Status: PROPOSED
// authority_effect: NONE
// Package outbox defines the in-memory projection of the durable outbox log.
package outbox

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	cyerrors "github.com/XRastlinX/logos-formal/pkg/cyexchange/errors"
)

type State int

const (
	StateUnknown State = iota
	StateQueued
	StateTransmitting
	StateDelivered
	StateFailed
	StateOutcomeUnknown
)

type Clock func() time.Time

type Adapter struct {
	mu    sync.RWMutex
	store map[string]*Record
	clock Clock
}

type Record struct {
	Envelope *envelope.CyExchangeEnvelope
	State    State
	QueuedAt time.Time
}

func NewAdapter() *Adapter {
	return NewAdapterWithClock(time.Now)
}

func NewAdapterWithClock(clock Clock) *Adapter {
	if clock == nil {
		clock = time.Now
	}
	return &Adapter{store: make(map[string]*Record), clock: clock}
}

func (a *Adapter) Enqueue(env *envelope.CyExchangeEnvelope) error {
	if env == nil || env.EnvelopeID == "" {
		return fmt.Errorf("%w: invalid outbox envelope", cyerrors.ErrValidation)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.store[env.EnvelopeID]; exists {
		return fmt.Errorf("%w: envelope already queued", cyerrors.ErrReplay)
	}
	a.store[env.EnvelopeID] = &Record{
		Envelope: cloneEnvelope(env),
		State:    StateQueued,
		QueuedAt: a.clock().UTC(),
	}
	return nil
}

func (a *Adapter) Restore(record *Record) error {
	if record == nil || record.Envelope == nil || record.Envelope.EnvelopeID == "" ||
		record.QueuedAt.IsZero() || !ValidState(record.State) {
		return fmt.Errorf("%w: invalid restored outbox record", cyerrors.ErrValidation)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.store[record.Envelope.EnvelopeID]; exists {
		return fmt.Errorf("%w: duplicate restored outbox envelope", cyerrors.ErrReplay)
	}
	a.store[record.Envelope.EnvelopeID] = &Record{
		Envelope: cloneEnvelope(record.Envelope),
		State:    record.State,
		QueuedAt: record.QueuedAt.UTC(),
	}
	return nil
}

func (a *Adapter) MarkTransmitting(envelopeID string) error {
	return a.transition(envelopeID, StateTransmitting)
}

func (a *Adapter) MarkDelivered(envelopeID string) error {
	return a.transition(envelopeID, StateDelivered)
}

func (a *Adapter) MarkFailed(envelopeID string) error {
	return a.transition(envelopeID, StateFailed)
}

func (a *Adapter) MarkOutcomeUnknown(envelopeID string) error {
	return a.transition(envelopeID, StateOutcomeUnknown)
}

func (a *Adapter) Transition(envelopeID string, newState State) error {
	return a.transition(envelopeID, newState)
}

func (a *Adapter) transition(envelopeID string, newState State) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	record, exists := a.store[envelopeID]
	if !exists {
		return errors.New("record not found in outbox")
	}
	if !CanTransition(record.State, newState) {
		return fmt.Errorf(
			"%w: illegal outbox transition %s -> %s",
			cyerrors.ErrValidation,
			record.State,
			newState,
		)
	}
	record.State = newState
	return nil
}

func (a *Adapter) Get(envelopeID string) (*Record, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	record, exists := a.store[envelopeID]
	if !exists {
		return nil, errors.New("record not found")
	}
	return &Record{
		Envelope: cloneEnvelope(record.Envelope),
		State:    record.State,
		QueuedAt: record.QueuedAt,
	}, nil
}

func (a *Adapter) GetState(envelopeID string) (State, error) {
	record, err := a.Get(envelopeID)
	if err != nil {
		return StateUnknown, err
	}
	return record.State, nil
}

func (a *Adapter) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.store)
}

func CanTransition(from, to State) bool {
	switch from {
	case StateQueued:
		return to == StateTransmitting || to == StateFailed
	case StateTransmitting:
		return to == StateDelivered || to == StateFailed || to == StateOutcomeUnknown
	case StateOutcomeUnknown:
		return to == StateDelivered || to == StateFailed
	default:
		return false
	}
}

func ValidState(state State) bool {
	return state >= StateQueued && state <= StateOutcomeUnknown
}

func (state State) String() string {
	switch state {
	case StateQueued:
		return "QUEUED"
	case StateTransmitting:
		return "TRANSMITTING"
	case StateDelivered:
		return "DELIVERED"
	case StateFailed:
		return "FAILED"
	case StateOutcomeUnknown:
		return "OUTCOME_UNKNOWN"
	default:
		return "UNKNOWN"
	}
}

func cloneEnvelope(source *envelope.CyExchangeEnvelope) *envelope.CyExchangeEnvelope {
	if source == nil {
		return nil
	}
	copyValue := *source
	copyValue.PayloadBytes = append([]byte(nil), source.PayloadBytes...)
	copyValue.Lineage.ParentEnvelopeIDs = append(
		[]string(nil),
		source.Lineage.ParentEnvelopeIDs...,
	)
	return &copyValue
}
