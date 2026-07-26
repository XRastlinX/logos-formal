package main

import (
	"fmt"
)

// BleedOverOperator applies the β translation while enforcing the Active Canon invariants.
type BleedOverOperator struct{}

// Translate attempts to move a ClaimPacket into a TargetContext, producing a BoundaryObject.
func (op *BleedOverOperator) Translate(c ClaimPacket, target TargetContext, proposedTranslation string, proposedDelta string, proposedUncertainty float64) (*BoundaryObject, error) {
	fmt.Printf("[β Operator] Processing Translation of Claim: %s -> Context: %s\n", c.ID, target.FieldID)

	// INVARIANT 1: Provenance Conservation
	if target.RequiresProvenance && len(c.Provenance) == 0 {
		return nil, fmt.Errorf("SECURITY FAULT [BO-1]: Claim lacks provenance. Cannot translate into context requiring provenance")
	}

	// INVARIANT 2 & 3: Authority Non-Transfer and Effect Non-Creation
	// The membrane forces Authority and Effect to 0, regardless of the input.
	// Translation cannot mint authorization or produce an effect.
	safeAuthority := 0
	safeEffect := 0

	// INVARIANT 4: Uncertainty Non-Inflation
	// A summary may become shorter, but it may not become surer (Uncertainty cannot decrease).
	if proposedUncertainty < c.Uncertainty {
		return nil, fmt.Errorf("SECURITY FAULT [BO-3]: Uncertainty inflation detected. Proposed %v, but source was %v. Translation cannot invent certainty", proposedUncertainty, c.Uncertainty)
	}

	fmt.Printf("[β Operator] Translation safe. Constructing Boundary Object.\n")
	
	// Construct the Boundary Object
	bo := &BoundaryObject{
		TargetRepresentation: proposedTranslation,
		ProvenancePath:       append(c.Provenance, fmt.Sprintf("Translated by β into %s", target.FieldID)),
		SemanticDelta:        proposedDelta,
		Authority:            safeAuthority,
		Effect:               safeEffect,
		Uncertainty:          proposedUncertainty,
	}

	return bo, nil
}
