// Status: PROPOSED
// authority_effect: NONE
// Package recovery reconstructs local queue projections without executing handlers.
package recovery

import (
	"fmt"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/inbox"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/outbox"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/storage"
)

type Report struct {
	InboxRestored        int
	OutboxRestored       int
	ExpiredRejected      int
	InboxOutcomeUnknown  int
	OutboxOutcomeUnknown int
}

type Projection struct {
	Inbox  *inbox.Adapter
	Outbox *outbox.Adapter
	Report Report
}

func Recover(store storage.Store, now time.Time) (*Projection, error) {
	if store == nil || now.IsZero() {
		return nil, fmt.Errorf("store and recovery time are required")
	}
	clock := func() time.Time { return now }
	inboxAdapter := inbox.NewAdapterWithClock(clock)
	outboxAdapter := outbox.NewAdapterWithClock(clock)
	report := Report{}

	inboxRecords, err := store.GetRecoverableLog()
	if err != nil {
		return nil, fmt.Errorf("recover inbox: %w", err)
	}
	for _, record := range inboxRecords {
		state := record.State
		if now.After(record.ExpiresAt) &&
			state != inbox.StateObserveOnly &&
			state != inbox.StateRejected {
			if err := store.UpdateState(record.Envelope.EnvelopeID, inbox.StateRejected); err != nil {
				return nil, fmt.Errorf("reject expired inbox record: %w", err)
			}
			state = inbox.StateRejected
			report.ExpiredRejected++
		} else if state == inbox.StateEvaluating {
			if err := store.UpdateState(
				record.Envelope.EnvelopeID,
				inbox.StateOutcomeUnknown,
			); err != nil {
				return nil, fmt.Errorf("mark inbox outcome unknown: %w", err)
			}
			state = inbox.StateOutcomeUnknown
			report.InboxOutcomeUnknown++
		}
		if err := inboxAdapter.Restore(&inbox.Record{
			Envelope:  record.Envelope,
			State:     state,
			ExpiresAt: record.ExpiresAt,
		}); err != nil {
			return nil, fmt.Errorf("restore inbox projection: %w", err)
		}
		report.InboxRestored++
	}

	outboxRecords, err := store.GetRecoverableOutbox()
	if err != nil {
		return nil, fmt.Errorf("recover outbox: %w", err)
	}
	for _, record := range outboxRecords {
		state := record.State
		if state == outbox.StateTransmitting {
			if err := store.UpdateOutboxState(
				record.Envelope.EnvelopeID,
				outbox.StateOutcomeUnknown,
			); err != nil {
				return nil, fmt.Errorf("mark outbox outcome unknown: %w", err)
			}
			state = outbox.StateOutcomeUnknown
			report.OutboxOutcomeUnknown++
		}
		if err := outboxAdapter.Restore(&outbox.Record{
			Envelope: record.Envelope,
			State:    state,
			QueuedAt: record.QueuedAt,
		}); err != nil {
			return nil, fmt.Errorf("restore outbox projection: %w", err)
		}
		report.OutboxRestored++
	}

	return &Projection{
		Inbox:  inboxAdapter,
		Outbox: outboxAdapter,
		Report: report,
	}, nil
}
