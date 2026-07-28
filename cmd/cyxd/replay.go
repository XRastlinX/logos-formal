// Status: PROPOSED
// authority_effect: NONE
package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	cyerrors "github.com/XRastlinX/logos-formal/pkg/cyexchange/errors"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/storage"
)

// enforceReplayImmunity verifies that a newly arrived envelope does not duplicate
// an existing envelope ID in the durable store.
func enforceReplayImmunity(env *envelope.CyExchangeEnvelope, store storage.Store) error {
	_, _, err := store.GetEnvelope(env.EnvelopeID)
	if err == nil {
		// Found it. This is a replay attempt.
		log.Printf("[REJECT] Replay detected for Envelope ID: %s", env.EnvelopeID)
		return fmt.Errorf("%w: duplicate envelope %s", cyerrors.ErrReplay, env.EnvelopeID)
	}
	if errors.Is(err, cyerrors.ErrNotFound) {
		return nil
	}
	return fmt.Errorf("%w: replay lookup: %v", cyerrors.ErrDurability, err)
}
