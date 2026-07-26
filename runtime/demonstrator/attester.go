package main

import (
	"fmt"
	"time"
)

// Attester takes a validated artifact, binds it to the current runtime identity,
// and cryptographically seals it into Evidence.
// Capability: 010 (Interpretation/Measurement only)
type Attester struct {
	ID string
}

func (a *Attester) Seal(art Artifact, validatorID string) Evidence {
	root := HashArtifact(art)
	fmt.Printf("[Attester] Sealing artifact %s... Generating PIC-4 Evidence.\n", root[:8])

	return Evidence{
		ArtifactRoot: root,
		ValidatorID:  validatorID,
		AttesterID:   a.ID,
		Timestamp:    time.Now().UTC().Format(time.RFC3339),
	}
}
