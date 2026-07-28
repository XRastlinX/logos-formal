// Status: PROPOSED
// authority_effect: NONE
package errors

import "errors"

var (
	ErrTransport  = errors.New("cyexchange: transport failure")
	ErrValidation = errors.New("cyexchange: schema or invariant validation failed")
	ErrSignature  = errors.New("cyexchange: ed25519 signature verification failed")
	ErrPermit     = errors.New("cyexchange: sovereign apply permit rejected")
	ErrReplay     = errors.New("cyexchange: deduplication or replay detected")
	ErrExpiration = errors.New("cyexchange: ttl or permit bounds expired")
	ErrLineage    = errors.New("cyexchange: epistemic lineage broken")
	ErrGovernance = errors.New("cyexchange: illegal cubed bit geometry")
	ErrNotFound   = errors.New("cyexchange: record not found")
	ErrDurability = errors.New("cyexchange: durable storage failure")
	ErrCorrupt    = errors.New("cyexchange: durable log corrupt")
)
