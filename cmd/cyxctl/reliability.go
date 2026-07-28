// Status: PROPOSED
// authority_effect: NONE
package main

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/discovery"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	cyerrors "github.com/XRastlinX/logos-formal/pkg/cyexchange/errors"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/recovery"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/storage"
)

// ReliabilityCommand runs bounded, local conformance checks. It reports PASS
// only after the corresponding operation has actually completed.
func ReliabilityCommand() error {
	fmt.Println("[PROPOSED] Running CyExchange V0.3 Reliability Suite")

	tempDir, err := os.MkdirTemp("", "cyexchange-reliability-*")
	if err != nil {
		return fmt.Errorf("create isolated test directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := storage.NewSQLiteStore(filepath.Join(tempDir, "suite.db"))
	if err != nil {
		return err
	}
	defer store.Close()

	payload := []byte(`{"request":"observe"}`)
	payloadHash, err := envelope.HashPayload(payload)
	if err != nil {
		return err
	}
	root := &envelope.CyExchangeEnvelope{
		EnvelopeID:       "reliability-root",
		PayloadBytes:     payload,
		PayloadHash:      payloadHash,
		PayloadSchemaURI: "urn:cyexchange:test",
		RequiredState:    "010",
	}
	now := time.Now().UTC()
	if err := store.InsertEnvelope(root, now.Add(time.Hour)); err != nil {
		return fmt.Errorf("insert durable root: %w", err)
	}

	projection, err := recovery.Recover(store, now)
	if err != nil {
		return fmt.Errorf("recover durable inbox: %w", err)
	}
	if projection.Report.InboxRestored != 1 || projection.Inbox.Len() != 1 {
		return fmt.Errorf("recovery count mismatch")
	}
	fmt.Println("[PASS] Crash recovery reconstructed one durable inbox record.")

	if err := store.InsertEnvelope(root, now.Add(time.Hour)); !errors.Is(err, cyerrors.ErrReplay) {
		return fmt.Errorf("duplicate insert did not return ErrReplay: %v", err)
	}
	fmt.Println("[PASS] SQLite uniqueness rejected a duplicate envelope ID.")

	forged := &envelope.CyExchangeEnvelope{
		EnvelopeID:       "reliability-child",
		PayloadBytes:     payload,
		PayloadHash:      payloadHash,
		PayloadSchemaURI: "urn:cyexchange:test",
		RequiredState:    "010",
		Lineage: envelope.Lineage{
			ParentEnvelopeIDs: []string{"missing-parent"},
			GenerationDepth:   1,
			LinkHash:          strings.Repeat("0", 64),
		},
	}
	if err := storage.VerifyLineage(forged, store); !errors.Is(err, cyerrors.ErrLineage) {
		return fmt.Errorf("forged lineage did not return ErrLineage: %v", err)
	}
	fmt.Println("[PASS] Missing-parent lineage was rejected.")

	prober := discovery.NewProber()
	before := httptest.NewRecorder()
	prober.Handler().ServeHTTP(before, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if before.Code != http.StatusServiceUnavailable {
		return fmt.Errorf("pre-recovery readiness returned %d", before.Code)
	}
	prober.MarkReady()
	after := httptest.NewRecorder()
	prober.Handler().ServeHTTP(after, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if after.Code != http.StatusOK {
		return fmt.Errorf("post-recovery readiness returned %d", after.Code)
	}
	fmt.Println("[PASS] Readiness remained closed until explicitly marked ready.")
	fmt.Println("\n[PROPOSED] Reliability Suite complete. No 101 boundaries were crossed.")
	return nil
}
