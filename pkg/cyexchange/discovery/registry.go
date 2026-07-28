// Status: PROPOSED
// authority_effect: NONE
package discovery

import (
	"fmt"
	"sync"

	"github.com/XRastlinX/logos-formal/pkg/cyexchange/envelope"
)

// Peer represents a discovered CyExchange node.
type Peer struct {
	NodeCard *envelope.NodeCard
	Address  string // e.g., "192.168.1.100:8080"
}

// Registry manages the collection of discovered nodes on the local link.
// It is non-authoritative; knowing a peer exists does not grant access.
type Registry struct {
	mu    sync.RWMutex
	peers map[string]*Peer
}

func NewRegistry() *Registry {
	return &Registry{
		peers: make(map[string]*Peer),
	}
}

// AddDiscovered registers a peer found via mDNS.
func (r *Registry) AddDiscovered(card *envelope.NodeCard, address string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.peers[card.DID] = &Peer{
		NodeCard: card,
		Address:  address,
	}
}

// Lookup attempts to resolve a DID to a network address.
func (r *Registry) Lookup(did string) (*Peer, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peer, exists := r.peers[did]
	if !exists {
		return nil, fmt.Errorf("node DID %s not found in registry", did)
	}
	return peer, nil
}
