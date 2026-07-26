package main

// ClaimPacket represents an extracted claim from a source field.
type ClaimPacket struct {
	ID            string
	Proposition   string
	SourceContext string
	Authority     int     // 0 = none, 1 = authoritative effect
	Effect        int     // 0 = none, 1 = direct actuation
	Uncertainty   float64 // 0.0 (Absolute Certainty) to 1.0 (Total Unknown)
	Provenance    []string
}

// TargetContext represents the ontological assumptions of the destination field.
type TargetContext struct {
	FieldID            string
	AllowedVocab       []string
	RequiresProvenance bool
}

// BoundaryObject is the result of applying the Bleed-Over Operator β.
type BoundaryObject struct {
	TargetRepresentation string
	ProvenancePath       []string
	SemanticDelta        string
	Authority            int
	Effect               int
	Uncertainty          float64
}
