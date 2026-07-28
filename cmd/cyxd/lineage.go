// Status: PROPOSED
// authority_effect: NONE
package main

import (
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/storage"
)

// trackLineage verifies the declared lineage against exact parent records
// already present in the durable store.
func trackLineage(env *envelope.CyExchangeEnvelope, store storage.Store) error {
	return storage.VerifyLineage(env, store)
}
