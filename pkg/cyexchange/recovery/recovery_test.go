// Status: PROPOSED
// authority_effect: NONE
package recovery

import (
	"testing"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/inbox"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/outbox"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/storage"
)

func TestRecoverMarksInterruptedWorkOutcomeUnknown(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	store := storage.NewMemoryStoreWithClock(func() time.Time { return now })
	inbound := &envelope.CyExchangeEnvelope{EnvelopeID: "inbound"}
	outbound := &envelope.CyExchangeEnvelope{EnvelopeID: "outbound"}

	if err := store.InsertEnvelope(inbound, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateState(inbound.EnvelopeID, inbox.StateEvaluating); err != nil {
		t.Fatal(err)
	}
	if err := store.EnqueueOutbox(outbound, now.Add(-time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateOutboxState(outbound.EnvelopeID, outbox.StateTransmitting); err != nil {
		t.Fatal(err)
	}

	projection, err := Recover(store, now)
	if err != nil {
		t.Fatal(err)
	}
	inboxState, err := projection.Inbox.GetState(inbound.EnvelopeID)
	if err != nil {
		t.Fatal(err)
	}
	outboxState, err := projection.Outbox.GetState(outbound.EnvelopeID)
	if err != nil {
		t.Fatal(err)
	}
	if inboxState != inbox.StateOutcomeUnknown ||
		outboxState != outbox.StateOutcomeUnknown {
		t.Fatalf("unexpected recovered states: %s, %s", inboxState, outboxState)
	}
	if projection.Report.InboxOutcomeUnknown != 1 ||
		projection.Report.OutboxOutcomeUnknown != 1 {
		t.Fatalf("unexpected report: %#v", projection.Report)
	}
}

func TestRecoverRejectsExpiredActiveInboxRecord(t *testing.T) {
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	store := storage.NewMemoryStoreWithClock(func() time.Time { return now })
	env := &envelope.CyExchangeEnvelope{EnvelopeID: "expired"}
	if err := store.InsertEnvelope(env, now.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	projection, err := Recover(store, now)
	if err != nil {
		t.Fatal(err)
	}
	state, err := projection.Inbox.GetState(env.EnvelopeID)
	if err != nil {
		t.Fatal(err)
	}
	if state != inbox.StateRejected || projection.Report.ExpiredRejected != 1 {
		t.Fatalf("unexpected expired recovery: %s, %#v", state, projection.Report)
	}
}
