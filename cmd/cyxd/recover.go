// Status: PROPOSED
// authority_effect: NONE
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/inbox"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/recovery"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/storage"
)

// recoverInbox reconstructs the monotonic inbox from the durable SQLite log.
// It explicitly halts at the epistemic boundary; no 101 commands are executed during replay.
func recoverInbox(store storage.Store) (*inbox.Adapter, error) {
	log.Println("[PROPOSED] Initiating V0.3 crash recovery...")
	projection, err := recovery.Recover(store, time.Now().UTC())
	if err != nil {
		return nil, fmt.Errorf("recovery failed: %w", err)
	}
	log.Printf(
		"[PROPOSED] Recovered %d envelopes; %d uncertain outcomes recorded.",
		projection.Report.InboxRestored,
		projection.Report.InboxOutcomeUnknown,
	)
	return projection.Inbox, nil
}
