// Status: PROPOSED
// authority_effect: NONE
// Package discovery provides LAN mDNS advertisement and registry tracking.
package discovery

import (
	"log"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
)

// Advertiser simulates an mDNS / DNS-SD service advertiser.
type Advertiser struct {
	nodeCard *envelope.NodeCard
	port     string
}

// NewAdvertiser initializes the local node's advertisement service.
func NewAdvertiser(card *envelope.NodeCard, port string) *Advertiser {
	return &Advertiser{
		nodeCard: card,
		port:     port,
	}
}

// Broadcast simulates broadcasting the NodeCard via mDNS.
// This is strictly a 010 discovery operation; it grants no authority to connect.
func (a *Advertiser) Broadcast() error {
	log.Printf("[PROPOSED mDNS] Broadcasting CyExchange NodeCard: %s at port %s", a.nodeCard.DID, a.port)
	return nil
}

// Stop simulates halting the mDNS advertisement.
func (a *Advertiser) Stop() {
	log.Printf("[PROPOSED mDNS] Halting advertisement for %s", a.nodeCard.DID)
}
