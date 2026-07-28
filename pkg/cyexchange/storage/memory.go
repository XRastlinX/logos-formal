// Status: PROPOSED
// authority_effect: NONE
// Package storage defines append-only durable queue contracts for CyExchange.
package storage

import (
	"fmt"
	"sync"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	cyerrors "github.com/XRastlinX/logos-formal/pkg/cyexchange/errors"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/inbox"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/outbox"
)

type Clock func() time.Time

type Store interface {
	InsertEnvelope(env *envelope.CyExchangeEnvelope, expiresAt time.Time) error
	UpdateState(envelopeID string, newState inbox.State) error
	GetEnvelope(envelopeID string) (*envelope.CyExchangeEnvelope, inbox.State, error)
	GetRecoverableLog() ([]*inbox.Record, error)

	EnqueueOutbox(env *envelope.CyExchangeEnvelope, queuedAt time.Time) error
	UpdateOutboxState(envelopeID string, newState outbox.State) error
	GetOutbox(envelopeID string) (*envelope.CyExchangeEnvelope, outbox.State, error)
	GetRecoverableOutbox() ([]*outbox.Record, error)

	EventCount(queueKind, envelopeID string) (int, error)
	Close() error
}

type MemoryStore struct {
	mu            sync.RWMutex
	clock         Clock
	inboxRecords  map[string]*inbox.Record
	inboxEvents   map[string][]inbox.State
	outboxRecords map[string]*outbox.Record
	outboxEvents  map[string][]outbox.State
}

func NewMemoryStore() *MemoryStore {
	return NewMemoryStoreWithClock(time.Now)
}

func NewMemoryStoreWithClock(clock Clock) *MemoryStore {
	if clock == nil {
		clock = time.Now
	}
	return &MemoryStore{
		clock:         clock,
		inboxRecords:  make(map[string]*inbox.Record),
		inboxEvents:   make(map[string][]inbox.State),
		outboxRecords: make(map[string]*outbox.Record),
		outboxEvents:  make(map[string][]outbox.State),
	}
}

func (m *MemoryStore) InsertEnvelope(
	env *envelope.CyExchangeEnvelope,
	expiresAt time.Time,
) error {
	if env == nil || env.EnvelopeID == "" || expiresAt.IsZero() {
		return fmt.Errorf("%w: invalid inbox insert", cyerrors.ErrValidation)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.inboxRecords[env.EnvelopeID]; exists {
		return fmt.Errorf("%w: duplicate envelope %s", cyerrors.ErrReplay, env.EnvelopeID)
	}
	m.inboxRecords[env.EnvelopeID] = &inbox.Record{
		Envelope:  cloneEnvelope(env),
		State:     inbox.StateReceived,
		ExpiresAt: expiresAt.UTC(),
	}
	m.inboxEvents[env.EnvelopeID] = []inbox.State{inbox.StateReceived}
	return nil
}

func (m *MemoryStore) UpdateState(envelopeID string, newState inbox.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	record, exists := m.inboxRecords[envelopeID]
	if !exists {
		return cyerrors.ErrNotFound
	}
	if m.clock().After(record.ExpiresAt) && newState != inbox.StateRejected {
		return cyerrors.ErrExpiration
	}
	if !inbox.CanTransition(record.State, newState) {
		return fmt.Errorf(
			"%w: inbox %s -> %s",
			cyerrors.ErrValidation,
			record.State,
			newState,
		)
	}
	record.State = newState
	m.inboxEvents[envelopeID] = append(m.inboxEvents[envelopeID], newState)
	return nil
}

func (m *MemoryStore) GetEnvelope(
	envelopeID string,
) (*envelope.CyExchangeEnvelope, inbox.State, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	record, exists := m.inboxRecords[envelopeID]
	if !exists {
		return nil, inbox.StateUnknown, cyerrors.ErrNotFound
	}
	return cloneEnvelope(record.Envelope), record.State, nil
}

func (m *MemoryStore) GetRecoverableLog() ([]*inbox.Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	records := make([]*inbox.Record, 0, len(m.inboxRecords))
	for _, record := range m.inboxRecords {
		records = append(records, &inbox.Record{
			Envelope:  cloneEnvelope(record.Envelope),
			State:     record.State,
			ExpiresAt: record.ExpiresAt,
		})
	}
	sortInboxRecords(records)
	return records, nil
}

func (m *MemoryStore) EnqueueOutbox(
	env *envelope.CyExchangeEnvelope,
	queuedAt time.Time,
) error {
	if env == nil || env.EnvelopeID == "" || queuedAt.IsZero() {
		return fmt.Errorf("%w: invalid outbox insert", cyerrors.ErrValidation)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.outboxRecords[env.EnvelopeID]; exists {
		return fmt.Errorf("%w: duplicate outbox envelope %s", cyerrors.ErrReplay, env.EnvelopeID)
	}
	m.outboxRecords[env.EnvelopeID] = &outbox.Record{
		Envelope: cloneEnvelope(env),
		State:    outbox.StateQueued,
		QueuedAt: queuedAt.UTC(),
	}
	m.outboxEvents[env.EnvelopeID] = []outbox.State{outbox.StateQueued}
	return nil
}

func (m *MemoryStore) UpdateOutboxState(
	envelopeID string,
	newState outbox.State,
) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	record, exists := m.outboxRecords[envelopeID]
	if !exists {
		return cyerrors.ErrNotFound
	}
	if !outbox.CanTransition(record.State, newState) {
		return fmt.Errorf(
			"%w: outbox %s -> %s",
			cyerrors.ErrValidation,
			record.State,
			newState,
		)
	}
	record.State = newState
	m.outboxEvents[envelopeID] = append(m.outboxEvents[envelopeID], newState)
	return nil
}

func (m *MemoryStore) GetOutbox(
	envelopeID string,
) (*envelope.CyExchangeEnvelope, outbox.State, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	record, exists := m.outboxRecords[envelopeID]
	if !exists {
		return nil, outbox.StateUnknown, cyerrors.ErrNotFound
	}
	return cloneEnvelope(record.Envelope), record.State, nil
}

func (m *MemoryStore) GetRecoverableOutbox() ([]*outbox.Record, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	records := make([]*outbox.Record, 0, len(m.outboxRecords))
	for _, record := range m.outboxRecords {
		records = append(records, &outbox.Record{
			Envelope: cloneEnvelope(record.Envelope),
			State:    record.State,
			QueuedAt: record.QueuedAt,
		})
	}
	sortOutboxRecords(records)
	return records, nil
}

func (m *MemoryStore) EventCount(queueKind, envelopeID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	switch queueKind {
	case "INBOX":
		events, exists := m.inboxEvents[envelopeID]
		if !exists {
			return 0, cyerrors.ErrNotFound
		}
		return len(events), nil
	case "OUTBOX":
		events, exists := m.outboxEvents[envelopeID]
		if !exists {
			return 0, cyerrors.ErrNotFound
		}
		return len(events), nil
	default:
		return 0, fmt.Errorf("%w: unknown queue kind", cyerrors.ErrValidation)
	}
}

func (m *MemoryStore) Close() error {
	return nil
}
