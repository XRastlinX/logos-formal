// Status: PROPOSED
// authority_effect: NONE
package storage

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	cyerrors "github.com/XRastlinX/logos-formal/pkg/cyexchange/errors"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/inbox"
)

func testEnvelope(id string) *envelope.CyExchangeEnvelope {
	return &envelope.CyExchangeEnvelope{
		EnvelopeID:       id,
		PayloadBytes:     []byte(`{"request":"observe"}`),
		PayloadHash:      strings.Repeat("a", 64),
		PayloadSchemaURI: "urn:cyexchange:test",
		RequiredState:    "010",
	}
}

func TestSQLiteStorePersistsMonotonicHistory(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "queue.db")
	now := time.Date(2026, 7, 28, 12, 0, 0, 0, time.UTC)
	store, err := NewSQLiteStoreWithClock(dbPath, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}

	env := testEnvelope("persisted")
	if err := store.InsertEnvelope(env, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	if err := store.InsertEnvelope(env, now.Add(time.Hour)); !errors.Is(err, cyerrors.ErrReplay) {
		t.Fatalf("expected ErrReplay, got %v", err)
	}
	if err := store.UpdateState(env.EnvelopeID, inbox.StateEvaluating); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateState(env.EnvelopeID, inbox.StateObserveOnly); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateState(env.EnvelopeID, inbox.StateReceived); err == nil {
		t.Fatal("expected backward transition rejection")
	}
	if count, err := store.EventCount(queueInbox, env.EnvelopeID); err != nil || count != 3 {
		t.Fatalf("event count = %d, err = %v", count, err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}

	reopened, err := NewSQLiteStoreWithClock(dbPath, func() time.Time { return now })
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	records, err := reopened.GetRecoverableLog()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].State != inbox.StateObserveOnly {
		t.Fatalf("unexpected recovered records: %#v", records)
	}
}

func TestVerifyLineageBindsExistingParents(t *testing.T) {
	store := NewMemoryStore()
	now := time.Now().UTC()
	parent := testEnvelope("parent")
	parent.PayloadHash = strings.Repeat("b", 64)
	if err := store.InsertEnvelope(parent, now.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}

	child := testEnvelope("child")
	child.PayloadHash = strings.Repeat("c", 64)
	child.Lineage = envelope.Lineage{
		ParentEnvelopeIDs: []string{parent.EnvelopeID},
		GenerationDepth:   1,
		LinkHash: ComputeLineageLink(
			child.EnvelopeID,
			child.PayloadHash,
			[]ParentBinding{{
				EnvelopeID:  parent.EnvelopeID,
				PayloadHash: parent.PayloadHash,
				Depth:       0,
			}},
		),
	}
	if err := VerifyLineage(child, store); err != nil {
		t.Fatalf("expected valid lineage, got %v", err)
	}
	child.Lineage.LinkHash = strings.Repeat("0", 64)
	if err := VerifyLineage(child, store); !errors.Is(err, cyerrors.ErrLineage) {
		t.Fatalf("expected ErrLineage, got %v", err)
	}
}
