// Status: PROPOSED
// authority_effect: NONE
// Package inbox defines the in-memory projection of the durable inbox log.
package inbox

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
	StateReceived
	StateEvaluating
	StateObserveOnly
	StateEffectRequested
	StateApplied
	StateRejected
	StateOutcomeUnknown
)

type Clock func() time.Time

type Adapter struct {
	mu    sync.RWMutex
	store map[string]*Record
	clock Clock
}

type Record struct {
	Envelope  *envelope.CyExchangeEnvelope
	State     State
	ExpiresAt time.Time
}

func NewAdapter() *Adapter {
	return NewAdapterWithClock(time.Now)
}

func NewAdapterWithClock(clock Clock) *Adapter {
	if clock == nil {
		clock = time.Now
	}
	return &Adapter{
		store: make(map[string]*Record),
		clock: clock,
	}
}

func (a *Adapter) Ingest(env *envelope.CyExchangeEnvelope, ttl time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("%w: inbox TTL must be positive", cyerrors.ErrValidation)
	}
	return a.IngestUntil(env, a.clock().Add(ttl))
}

func (a *Adapter) IngestUntil(env *envelope.CyExchangeEnvelope, expiresAt time.Time) error {
	if env == nil || env.EnvelopeID == "" || expiresAt.IsZero() {
		return fmt.Errorf("%w: invalid inbox record", cyerrors.ErrValidation)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.store[env.EnvelopeID]; exists {
		return fmt.Errorf("%w: envelope %s already exists", cyerrors.ErrReplay, env.EnvelopeID)
	}
	a.store[env.EnvelopeID] = &Record{
		Envelope:  cloneEnvelope(env),
		State:     StateReceived,
		ExpiresAt: expiresAt.UTC(),
	}
	return nil
}

// Restore installs an already validated durable projection without running a
// handler or replaying a transition. It is the only crash-recovery entrypoint.
func (a *Adapter) Restore(record *Record) error {
	if record == nil || record.Envelope == nil || record.Envelope.EnvelopeID == "" ||
		record.ExpiresAt.IsZero() || !ValidState(record.State) {
		return fmt.Errorf("%w: invalid restored inbox record", cyerrors.ErrValidation)
	}
	if record.State == StateEffectRequested || record.State == StateApplied {
		return fmt.Errorf("%w: effect state cannot be restored into 010 inbox", cyerrors.ErrGovernance)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.store[record.Envelope.EnvelopeID]; exists {
		return fmt.Errorf("%w: duplicate restored envelope", cyerrors.ErrReplay)
	}
	a.store[record.Envelope.EnvelopeID] = &Record{
		Envelope:  cloneEnvelope(record.Envelope),
		State:     record.State,
		ExpiresAt: record.ExpiresAt.UTC(),
	}
	return nil
}

func (a *Adapter) Transition(envelopeID string, newState State) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	record, exists := a.store[envelopeID]
	if !exists {
		return fmt.Errorf("record not found: %s", envelopeID)
	}
	if a.clock().After(record.ExpiresAt) {
		return fmt.Errorf("%w: transition attempted on expired envelope", cyerrors.ErrExpiration)
	}
	if !CanTransition(record.State, newState) {
		return fmt.Errorf(
			"%w: illegal inbox transition %s -> %s",
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
		Envelope:  cloneEnvelope(record.Envelope),
		State:     record.State,
		ExpiresAt: record.ExpiresAt,
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
	case StateReceived:
		return to == StateEvaluating || to == StateRejected
	case StateEvaluating:
		return to == StateObserveOnly || to == StateRejected || to == StateOutcomeUnknown
	case StateOutcomeUnknown:
		return to == StateObserveOnly || to == StateRejected
	default:
		return false
	}
}

func ValidState(state State) bool {
	return state >= StateReceived && state <= StateOutcomeUnknown
}

func (state State) String() string {
	switch state {
	case StateReceived:
		return "RECEIVED"
	case StateEvaluating:
		return "EVALUATING"
	case StateObserveOnly:
		return "OBSERVE_ONLY"
	case StateEffectRequested:
		return "EFFECT_REQUESTED"
	case StateApplied:
		return "APPLIED"
	case StateRejected:
		return "REJECTED"
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
