// Status: PROPOSED
// authority_effect: NONE
package storage

import (
	"sort"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/inbox"
	"github.com/XRastlinX/logos-formal/pkg/cyexchange/outbox"
)

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

func sortInboxRecords(records []*inbox.Record) {
	sort.Slice(records, func(i, j int) bool {
		return records[i].Envelope.EnvelopeID < records[j].Envelope.EnvelopeID
	})
}

func sortOutboxRecords(records []*outbox.Record) {
	sort.Slice(records, func(i, j int) bool {
		return records[i].Envelope.EnvelopeID < records[j].Envelope.EnvelopeID
	})
}
